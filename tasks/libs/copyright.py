#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import datetime
import re
import subprocess
import sys
from pathlib import Path, PurePosixPath

GLOB_PATTERN = "**/*.go"

COPYRIGHT_HEADER = f"""
// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright {datetime.datetime.now().year}-present Datadog, Inc.
""".strip()

COPYRIGHT_REGEX = [
    r'^// Unless explicitly stated otherwise all files in this repository are licensed$',
    r'^// under the Apache License Version 2.0\.$',
    r'^// This product includes software developed at Datadog \(https://www\.[Dd]atadoghq\.com/\)\.$',
    r'^// Copyright 20[1-3][0-9]-([Pp]resent|20[1-3][0-9]) Datadog, (Inc|Inmetrics)\.$',
]

# These path patterns are excluded from checks
PATH_EXCLUSION_REGEX = [
    # These are auto-generated files but without headers to indicate it
    '/pkg/clusteragent/custommetrics/api/generated/',
    '/pkg/proto/msgpgo/.*_gen(_test){,1}.go',
    '/pkg/process/events/model/.*_gen.go',
    '/pkg/remoteconfig/state/products/apmsampling/.*_gen(_test){,1}.go',
    '/pkg/security/probe/accessors.go',
    '/pkg/security/security_profile/dump/activity_dump_easyjson.go',
    '/pkg/security/probe/custom_events_easyjson.go',
    '/pkg/security/serializers/json/serializers_easyjson.go',
    '/pkg/security/probe/dump/.*_gen(_test){,1}.go',
    '/pkg/security/secl/model/.*_gen(_test){,1}.go',
    '/pkg/security/secl/model/accessors.go',
    '/pkg/trace/pb/.*_gen(_test){,1}.go',
    # These are files that we should not add our copyright to
    '/internal/patch/grpc-go-insecure/',
    '/internal/patch/logr/funcr/funcr(_test){,1}.go',
    '/internal/patch/logr/funcr/internal/logr/',
    '/internal/third_party/golang/',
    '/internal/third_party/kubernetes/',
    '/pkg/collector/corechecks/cluster/ksm/customresources/utils.go',
]

# These header matchers skip enforcement of the rules if found in the first
# line of the file
HEADER_EXCLUSION_REGEX = [
    '^// Code generated ',
    '^//go:generate ',
    '^// AUTOGENERATED FILE: ',
    '^// Copyright.* OpenTelemetry Authors',
    '^// Copyright.* The Go Authors',
    '^// This file includes software developed at CoreOS',
    '^// Copyright 2017 Kinvolk',
    '^// Copyright 2021 The Vitess Authors.',
]


COMPILED_COPYRIGHT_REGEX = [re.compile(regex, re.UNICODE) for regex in COPYRIGHT_REGEX]
COMPILED_PATH_EXCLUSION_REGEX = [re.compile(regex, re.UNICODE) for regex in PATH_EXCLUSION_REGEX]
COMPILED_HEADER_EXCLUSION_REGEX = [re.compile(regex, re.UNICODE) for regex in HEADER_EXCLUSION_REGEX]


class LintFailure(Exception):
    pass


class CopyrightLinter:
    """
    This class is used to enforce copyright headers on specified file patterns
    """

    def __init__(self, debug=False):
        self._debug = debug

    @staticmethod
    def _get_repo_dir():
        script_dir = PurePosixPath(__file__).parent

        repo_dir = (
            subprocess.check_output(
                ['git', 'rev-parse', '--show-toplevel'],
                cwd=script_dir,
            )
            .decode(sys.stdout.encoding)
            .strip()
        )

        return PurePosixPath(repo_dir)

    @staticmethod
    def _is_excluded_path(filepath, exclude_matchers):
        for matcher in exclude_matchers:
            if re.search(matcher, filepath.as_posix()):
                return True

        return False

    @staticmethod
    def _get_matching_files(root_dir, glob_pattern, exclude=None):
        if exclude is None:
            exclude = []

        # Glob is a generator so we have to do the counting ourselves
        all_matching_files_cnt = 0

        filtered_files = []
        for filepath in Path(root_dir).glob(glob_pattern):
            all_matching_files_cnt += 1
            if not CopyrightLinter._is_excluded_path(filepath, exclude):
                filtered_files.append(filepath)

        excluded_files_cnt = all_matching_files_cnt - len(filtered_files)
        print(f"[INFO] Excluding {excluded_files_cnt} files based on path filters!")

        return sorted(filtered_files)

    @staticmethod
    def _get_header(filepath):
        header = []
        with open(filepath, "r", encoding="utf-8") as file_obj:
            # We expect a specific header format which should be 4 lines
            for _ in range(4):
                header.append(file_obj.readline().strip())

        return header

    @staticmethod
    def _is_excluded_header(header, exclude=None):
        if exclude is None:
            exclude = []

        for matcher in exclude:
            if re.search(matcher, header[0]):
                return True

        return False

    def _has_copyright(self, filepath):
        header = CopyrightLinter._get_header(filepath)
        if header is None:
            print("[WARN] Mismatch found! Could not find any content in file!")
            return False

        if len(header) > 0 and CopyrightLinter._is_excluded_header(header, exclude=COMPILED_HEADER_EXCLUSION_REGEX):
            if self._debug:
                print(f"[INFO] Excluding {filepath} based on header '{header[0]}'")
            return True

        if len(header) <= 3:
            print("[WARN] Mismatch found! File too small for header stanza!")
            return False

        for line_idx, matcher in enumerate(COMPILED_COPYRIGHT_REGEX):
            if not re.match(matcher, header[line_idx]):
                print(
                    f"[WARN] Mismatch found! Expected '{COPYRIGHT_REGEX[line_idx]}' pattern but got '{header[line_idx]}'"
                )
                return False

        return True

    def _assert_copyrights(self, files):
        failing_files = []
        for filepath in files:
            if self._has_copyright(filepath):
                if self._debug:
                    print(f"[ OK ] {filepath}")

                continue

            print(f"[FAIL] {filepath}")
            failing_files.append(filepath)

        total_files = len(files)
        if failing_files:
            pct_failing = (len(failing_files) / total_files) * 100
            print()
            print(
                f"FAIL: There are {len(failing_files)} files out of "
                + f"{total_files} ({pct_failing:.2f}%) that are missing the proper copyright!"
            )

        return failing_files

    def _prepend_header(self, filepath, dry_run=True):
        with open(filepath, 'r+') as file_obj:
            existing_content = file_obj.read()

            if dry_run:
                return True

            file_obj.seek(0)
            new_content = COPYRIGHT_HEADER + "\n\n" + existing_content
            file_obj.write(new_content)

        # Verify result. A problem here is not benign so we stop the whole run.
        if not self._has_copyright(filepath):
            raise Exception(f"[ERROR] Header prepend failed to produce correct output for {filepath}!")

        return True

    @staticmethod
    def _is_build_header(line):
        return line.startswith("// +build ") or line.startswith("//+build ") or line.startswith("//go:build ")

    def _fix_file_header(self, filepath, dry_run=True):
        header = CopyrightLinter._get_header(filepath)

        # Empty file - ignore
        if len(header) < 1:
            return False

        # If the file starts with a comment and it's not a build comment,
        # there is likely a manual fix to the header needed
        if header[0].startswith("//") and not CopyrightLinter._is_build_header(header[0]):
            return False

        if dry_run:
            return True

        return self._prepend_header(filepath, dry_run=dry_run)

    def _fix(self, failing_files, dry_run=True):
        failing_files_cnt = len(failing_files)
        errors = []
        for idx, filepath in enumerate(failing_files):
            print(f"[INFO] ({idx+1:3d}/{failing_files_cnt:3}) Fixing '{filepath}'...")

            if not self._fix_file_header(filepath, dry_run=dry_run):
                error_message = f"'{filepath}' could not be fixed!"
                print(f"[WARN] ({idx+1:3d}/{failing_files_cnt:3}) {error_message}")
                errors.append(LintFailure(error_message))

        return errors

    def assert_compliance(self, fix=False, dry_run=True, files=None):
        """
        This method verifies that all named files have the expected copyright header.

        If files is not given, this method applies the GLOB_PATTERN to the root
        of the repository to determine the files to check.
        """
        if files is None:
            git_repo_dir = CopyrightLinter._get_repo_dir()

            if self._debug:
                print(f"[DEBG] Repo root: {git_repo_dir}")
                print(f"[DEBG] Finding all files in {git_repo_dir} matching '{GLOB_PATTERN}'...")

            files = CopyrightLinter._get_matching_files(
                git_repo_dir,
                GLOB_PATTERN,
                exclude=COMPILED_PATH_EXCLUSION_REGEX,
            )
            print(f"[INFO] Found {len(files)} files matching '{GLOB_PATTERN}'")

        failing_files = self._assert_copyrights(files)
        if len(failing_files) > 0:
            if not fix:
                print("CHECK: FAIL")
                raise LintFailure(
                    f"Copyright linting found {len(failing_files)} files that did not have the expected header!"
                )

            # If "fix=True", we will attempt to fix the failing files
            errors = self._fix(failing_files, dry_run=dry_run)
            if errors:
                raise LintFailure(f"Copyright linter was unable to fix {len(errors)}/{len(failing_files)} files!")

            return

        print("CHECK: OK")


if __name__ == '__main__':
    CopyrightLinter(debug=True).assert_compliance()
