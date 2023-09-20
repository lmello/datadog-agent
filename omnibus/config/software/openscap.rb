# Unless explicitly stated otherwise all files in this repository are licensed
# under the Apache License Version 2.0.
# This product includes software developed at Datadog (https:#www.datadoghq.com/).
# Copyright 2016-present Datadog, Inc.

require './lib/cmake.rb'

name 'openscap'
default_version '1.3.9'

license "LGPL-3.0-or-later"
license_file "COPYING"

version("1.3.9") { source sha256: "2d8450b6b6ef068991e1292cd3989e8a1d81f2bcda0a2644dcb2943c2de1a20d" }

ship_source_offer true

source url: "https://github.com/OpenSCAP/openscap/releases/download/#{version}/openscap-#{version}.tar.gz"

dependency 'apt'
dependency 'attr'
dependency 'bzip2'
dependency 'curl'
dependency 'dbus'
dependency 'libacl'
dependency 'libgcrypt'
dependency 'libselinux'
dependency 'libsepol'
dependency 'libxslt'
dependency 'libyaml'
dependency 'pcre2'
dependency 'popt'
dependency 'rpm'
dependency 'util-linux'
dependency 'xmlsec'

relative_path "openscap-#{version}"

build do
  env = with_standard_compiler_flags(with_embedded_path)

  # Pending fixes.
  patch source: "1001-Fix-deadlocks-reported-by-Coverity.patch", env: env
  patch source: "1002-Fix-memory-leaks-reported-by-Coverity.patch", env: env
  patch source: "1003-Fix-out-of-bounds-access-reported-by-Coverity.patch", env: env
  patch source: "1004-Fix-incorrect-expressions-reported-by-Coverity.patch", env: env
  patch source: "1005-Fix-uninitialized-variables-reported-by-Coverity.patch", env: env

  patch source: "get_results_from_session.patch", env: env # add a function to retrieve results from session
  patch source: "session_result_reset.patch", env: env # add a function to reset results from session
  patch source: "session_reset_syschar.patch", env: env # also reset system characteristics
  patch source: "session_reset_results.patch", env: env # also reset OVAL results
  patch source: "010_perlpm_install_fix.patch", env: env # fix build of perl bindings
  patch source: "dpkginfo-cacheconfig.patch", env: env # work around incomplete pkgcache path
  patch source: "dpkginfo-cache-no-open.patch", env: env # reduce memory footprint of dpkginfo probe
  patch source: "fsdev-ignore-host.patch", env: env # ignore /host directory in fsdev probe
  patch source: "systemd-dbus-address.patch", env: env # fix dbus address in systemd probe
  patch source: "rpm-verbosity-err.patch", env: env # decrease rpmlog verbosity level to ERR
  patch source: "session-print-syschar.patch", env: env # add a function to print system characteristics
  patch source: "memusage-cgroup.patch", env: env # consider cgroup when determining memory usage

  patch source: "oscap-io.patch", env: env # add new oscap-io tool

  env["CC"] = "/opt/gcc-#{ENV['GCC_VERSION']}/bin/gcc"
  env["CXX"] = "/opt/gcc-#{ENV['GCC_VERSION']}/bin/g++"
  env["CXXFLAGS"] += " -static-libstdc++ -std=c++11 -DDPKG_DATADIR=/usr/share/dpkg"

  cmake_build_dir = "#{project_dir}/build"
  cmake_options = [
    "-DENABLE_PERL=OFF",
    "-DENABLE_PYTHON3=OFF",
    "-DWITH_PCRE2=ON",
    "-DENABLE_TESTS=OFF",
    "-DACL_INCLUDE_DIR:PATH=#{install_dir}/embedded/include",
    "-DACL_LIBRARY:FILEPATH=#{install_dir}/embedded/lib/libacl.so",
    "-DAPTPKG_INCLUDE_DIR:PATH=#{install_dir}/embedded/include",
    "-DAPTPKG_LIBRARIES:FILEPATH=#{install_dir}/embedded/lib/libapt-pkg.so",
    "-DBLKID_INCLUDE_DIR:PATH=#{install_dir}/embedded/include",
    "-DBLKID_LIBRARY:FILEPATH=#{install_dir}/embedded/lib/libblkid.so",
    "-DBZIP2_INCLUDE_DIR:PATH=#{install_dir}/embedded/include",
    "-DBZIP2_LIBRARY_RELEASE:FILEPATH=#{install_dir}/embedded/lib/libbz2.so",
    "-DCURL_INCLUDE_DIR:PATH=#{install_dir}/embedded/include",
    "-DCURL_LIBRARY_RELEASE:FILEPATH=#{install_dir}/embedded/lib/libcurl.so",
    "-DDBUS_INCLUDE_DIR:PATH=#{install_dir}/embedded/include/dbus-1.0",
    "-DDBUS_LIBRARIES:FILEPATH=#{install_dir}/embedded/lib/libdbus-1.so",
    "-DGCRYPT_INCLUDE_DIR:PATH=#{install_dir}/embedded/include",
    "-DGCRYPT_LIBRARY:FILEPATH=#{install_dir}/embedded/lib/libgcrypt.so",
    "-DLIBXML2_INCLUDE_DIR:PATH=#{install_dir}/embedded/include/libxml2",
    "-DLIBXML2_LIBRARY:FILEPATH=#{install_dir}/embedded/lib/libxml2.so",
    "-DLIBXSLT_EXSLT_INCLUDE_DIR:PATH=#{install_dir}/embedded/include",
    "-DLIBXSLT_EXSLT_LIBRARY:FILEPATH=#{install_dir}/embedded/lib/libexslt.so",
    "-DLIBXSLT_INCLUDE_DIR:PATH=#{install_dir}/embedded/include",
    "-DLIBXSLT_LIBRARY:FILEPATH=#{install_dir}/embedded/lib/libxslt.so",
    "-DLIBYAML_INCLUDE_DIR:PATH=#{install_dir}/embedded/include",
    "-DLIBYAML_LIBRARY:FILEPATH=#{install_dir}/embedded/lib/libyaml.so",
    "-DOPENSSL_CRYPTO_LIBRARY:FILEPATH=#{install_dir}/embedded/lib/libcrypto.so",
    "-DOPENSSL_INCLUDE_DIR:PATH=#{install_dir}/embedded/include",
    "-DOPENSSL_SSL_LIBRARY:FILEPATH=#{install_dir}/embedded/lib/libssl.so",
    "-DPCRE2_INCLUDE_DIR:PATH=#{install_dir}/embedded/include",
    "-DPCRE2_LIBRARY:FILEPATH=#{install_dir}/embedded/lib/libpcre2-8.so",
    "-DPOPT_INCLUDE_DIR:PATH=#{install_dir}/embedded/include",
    "-DPOPT_LIBRARY:FILEPATH=#{install_dir}/embedded/lib/libpopt.so",
    "-DPYTHON_INCLUDE_DIR:PATH=#{install_dir}/embedded/include/python3.8",
    "-DPYTHON_LIBRARY:FILEPATH=#{install_dir}/embedded/lib/libpython3.8.so",
    "-DRPM_INCLUDE_DIR:PATH=#{install_dir}/embedded/include",
    "-DRPMIO_LIBRARY:FILEPATH=#{install_dir}/embedded/lib/librpmio.so",
    "-DRPM_LIBRARY:FILEPATH=#{install_dir}/embedded/lib/librpm.so",
    "-DSELINUX_INCLUDE_DIR:PATH=#{install_dir}/embedded/include",
    "-DSELINUX_LIBRARY:FILEPATH=#{install_dir}/embedded/lib/libselinux.so",
    "-DXMLSEC_INCLUDE_DIR:PATH=#{install_dir}/embedded/include/xmlsec1",
    "-DXMLSEC_LIBRARY:FILEPATH=#{install_dir}/embedded/lib/libxmlsec1.so",
    "-DXMLSEC_OPENSSL_LIBRARY:FILEPATH=#{install_dir}/embedded/lib/libxmlsec1-openssl.so",
  ]
  cmake(*cmake_options, env: env, cwd: cmake_build_dir, prefix: "#{install_dir}/embedded")

  # Remove OpenSCAP XML schemas, since they are not useful when XSD validation is disabled.
  delete "#{install_dir}/embedded/share/openscap/schemas"
end
