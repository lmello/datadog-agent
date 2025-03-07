// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2022-present Datadog, Inc.

package invocationlifecycle

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	pb "github.com/DataDog/datadog-agent/pkg/proto/pbgo/trace"
	"github.com/DataDog/datadog-agent/pkg/serverless/trace/inferredspan"

	"github.com/aws/aws-lambda-go/events"

	"github.com/DataDog/datadog-agent/pkg/serverless/trigger"
	"github.com/DataDog/datadog-agent/pkg/util/log"
)

const (
	tagFunctionTriggerEventSource    = "function_trigger.event_source"
	tagFunctionTriggerEventSourceArn = "function_trigger.event_source_arn"
)

func (lp *LifecycleProcessor) initFromAPIGatewayEvent(event events.APIGatewayProxyRequest, region string) {
	if !lp.DetectLambdaLibrary() && lp.InferredSpansEnabled {
		lp.GetInferredSpan().EnrichInferredSpanWithAPIGatewayRESTEvent(event)
	}

	lp.requestHandler.event = event
	lp.addTag(tagFunctionTriggerEventSource, apiGateway)
	lp.addTag(tagFunctionTriggerEventSourceArn, trigger.ExtractAPIGatewayEventARN(event, region))
	lp.addTags(trigger.GetTagsFromAPIGatewayEvent(event))
}

func (lp *LifecycleProcessor) initFromAPIGatewayV2Event(event events.APIGatewayV2HTTPRequest, region string) {
	if !lp.DetectLambdaLibrary() && lp.InferredSpansEnabled {
		lp.GetInferredSpan().EnrichInferredSpanWithAPIGatewayHTTPEvent(event)
	}

	lp.requestHandler.event = event
	lp.addTag(tagFunctionTriggerEventSource, apiGateway)
	lp.addTag(tagFunctionTriggerEventSourceArn, trigger.ExtractAPIGatewayV2EventARN(event, region))
	lp.addTags(trigger.GetTagsFromAPIGatewayV2HTTPRequest(event))
}

func (lp *LifecycleProcessor) initFromAPIGatewayWebsocketEvent(event events.APIGatewayWebsocketProxyRequest, region string) {
	if !lp.DetectLambdaLibrary() && lp.InferredSpansEnabled {
		lp.GetInferredSpan().EnrichInferredSpanWithAPIGatewayWebsocketEvent(event)
	}

	lp.requestHandler.event = event
	lp.addTag(tagFunctionTriggerEventSource, apiGateway)
	lp.addTag(tagFunctionTriggerEventSourceArn, trigger.ExtractAPIGatewayWebSocketEventARN(event, region))
}

func (lp *LifecycleProcessor) initFromAPIGatewayLambdaAuthorizerTokenEvent(event events.APIGatewayCustomAuthorizerRequest) {
	lp.requestHandler.event = event
	lp.addTag(tagFunctionTriggerEventSource, apiGateway)
	lp.addTag(tagFunctionTriggerEventSourceArn, trigger.ExtractAPIGatewayCustomAuthorizerEventARN(event))
	lp.addTags(trigger.GetTagsFromAPIGatewayCustomAuthorizerEvent(event))
}

func (lp *LifecycleProcessor) initFromAPIGatewayLambdaAuthorizerRequestParametersEvent(event events.APIGatewayCustomAuthorizerRequestTypeRequest) {
	lp.requestHandler.event = event
	lp.addTag(tagFunctionTriggerEventSource, apiGateway)
	lp.addTag(tagFunctionTriggerEventSourceArn, trigger.ExtractAPIGatewayCustomAuthorizerRequestTypeEventARN(event))
	lp.addTags(trigger.GetTagsFromAPIGatewayCustomAuthorizerRequestTypeEvent(event))
}

func (lp *LifecycleProcessor) initFromALBEvent(event events.ALBTargetGroupRequest) {
	lp.requestHandler.event = event
	lp.addTag(tagFunctionTriggerEventSource, applicationLoadBalancer)
	lp.addTag(tagFunctionTriggerEventSourceArn, trigger.ExtractAlbEventARN(event))
	lp.addTags(trigger.GetTagsFromALBTargetGroupRequest(event))
}

func (lp *LifecycleProcessor) initFromCloudWatchEvent(event events.CloudWatchEvent) {
	lp.requestHandler.event = event
	lp.addTag(tagFunctionTriggerEventSource, cloudwatchEvents)
	lp.addTag(tagFunctionTriggerEventSourceArn, trigger.ExtractCloudwatchEventARN(event))
}

func (lp *LifecycleProcessor) initFromCloudWatchLogsEvent(event events.CloudwatchLogsEvent, region string, accountID string) {
	arn, err := trigger.ExtractCloudwatchLogsEventARN(event, region, accountID)
	if err != nil {
		log.Debugf("Error parsing event ARN from cloudwatch logs event: %v", err)
		return
	}

	lp.requestHandler.event = event
	lp.addTag(tagFunctionTriggerEventSource, cloudwatchLogs)
	lp.addTag(tagFunctionTriggerEventSourceArn, arn)
}

func (lp *LifecycleProcessor) initFromDynamoDBStreamEvent(event events.DynamoDBEvent) {
	if !lp.DetectLambdaLibrary() && lp.InferredSpansEnabled {
		lp.GetInferredSpan().EnrichInferredSpanWithDynamoDBEvent(event)
	}

	lp.requestHandler.event = event
	lp.addTag(tagFunctionTriggerEventSource, dynamoDB)
	lp.addTag(tagFunctionTriggerEventSourceArn, trigger.ExtractDynamoDBStreamEventARN(event))
}

func (lp *LifecycleProcessor) initFromEventBridgeEvent(event inferredspan.EventBridgeEvent) {
	lp.requestHandler.event = event
	lp.addTag(tagFunctionTriggerEventSource, eventBridge)
	lp.addTag(tagFunctionTriggerEventSourceArn, event.Source)
}

func (lp *LifecycleProcessor) initFromKinesisStreamEvent(event events.KinesisEvent) {
	if !lp.DetectLambdaLibrary() && lp.InferredSpansEnabled {
		lp.GetInferredSpan().EnrichInferredSpanWithKinesisEvent(event)
	}

	lp.requestHandler.event = event
	lp.addTag(tagFunctionTriggerEventSource, kinesis)
	lp.addTag(tagFunctionTriggerEventSourceArn, trigger.ExtractKinesisStreamEventARN(event))
}

func (lp *LifecycleProcessor) initFromS3Event(event events.S3Event) {
	if !lp.DetectLambdaLibrary() && lp.InferredSpansEnabled {
		lp.GetInferredSpan().EnrichInferredSpanWithS3Event(event)
	}

	lp.requestHandler.event = event
	lp.addTag(tagFunctionTriggerEventSource, s3)
	lp.addTag(tagFunctionTriggerEventSourceArn, trigger.ExtractS3EventArn(event))
}

func (lp *LifecycleProcessor) initFromSNSEvent(event events.SNSEvent) {
	if !lp.DetectLambdaLibrary() && lp.InferredSpansEnabled {
		lp.GetInferredSpan().EnrichInferredSpanWithSNSEvent(event)
	}

	lp.requestHandler.event = event
	lp.addTag(tagFunctionTriggerEventSource, sns)
	lp.addTag(tagFunctionTriggerEventSourceArn, trigger.ExtractSNSEventArn(event))
}

func (lp *LifecycleProcessor) initFromSQSEvent(event events.SQSEvent) {
	if !lp.DetectLambdaLibrary() && lp.InferredSpansEnabled {
		lp.GetInferredSpan().EnrichInferredSpanWithSQSEvent(event)
	}

	lp.requestHandler.event = event
	lp.addTag(tagFunctionTriggerEventSource, sqs)
	lp.addTag(tagFunctionTriggerEventSourceArn, trigger.ExtractSQSEventARN(event))

	// test for SNS
	var snsEntity events.SNSEntity
	if err := json.Unmarshal([]byte(event.Records[0].Body), &snsEntity); err != nil {
		return
	}

	isSNS := strings.ToLower(snsEntity.Type) == "notification" && snsEntity.TopicArn != ""

	if !isSNS {
		return
	}

	// sns span
	lp.requestHandler.inferredSpans[1] = &inferredspan.InferredSpan{
		CurrentInvocationStartTime: time.Unix(lp.requestHandler.inferredSpans[0].Span.Start, 0),
		Span: &pb.Span{
			SpanID: inferredspan.GenerateSpanId(),
		},
	}

	var snsEvent events.SNSEvent
	snsEvent.Records = make([]events.SNSEventRecord, 1)
	snsEvent.Records[0].SNS = snsEntity

	lp.requestHandler.inferredSpans[1].EnrichInferredSpanWithSNSEvent(snsEvent)

	lp.requestHandler.inferredSpans[1].Span.Duration = lp.GetInferredSpan().Span.Start - lp.requestHandler.inferredSpans[1].Span.Start

}

func (lp *LifecycleProcessor) initFromLambdaFunctionURLEvent(event events.LambdaFunctionURLRequest, region string, accountID string, functionName string) {
	lp.requestHandler.event = event
	lp.addTag(tagFunctionTriggerEventSource, functionURL)
	lp.addTag(tagFunctionTriggerEventSourceArn, fmt.Sprintf("arn:aws:lambda:%v:%v:url:%v", region, accountID, functionName))
	lp.addTags(trigger.GetTagsFromLambdaFunctionURLRequest(event))
}
