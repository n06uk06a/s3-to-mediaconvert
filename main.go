package main

/*
 Launch AWS Elemental MediaConvert by S3 Event
*/
import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
)

// Cast string to pointer of string
func sToPtr(s string) *string {
	return &s
}

// Cast int64 to pointer of int64
func i64ToPtr(i int64) *int64 {
	return &i
}

// Get environment variables for setting
var destBucket = os.Getenv("DEST_BUCKET")
var queue = os.Getenv("QUEUE")
var role = os.Getenv("ROLE")

// Initialize MediaConvert session
var sess = session.Must(session.NewSession())
var config = aws.NewConfig()
var svc = mediaconvert.New(sess)

// Settings for MediaConvert
var settings = &mediaconvert.JobSettings{AdAvailOffset: i64ToPtr(0)}
var input = &mediaconvert.Input{
	AudioSelectors: map[string]*mediaconvert.AudioSelector{
		"Audio Selector 1": &mediaconvert.AudioSelector{
			DefaultSelection: sToPtr("DEFAULT"),
			Offset:           i64ToPtr(1),
			ProgramSelection: i64ToPtr(1)}},
	DeblockFilter:  sToPtr("DISABLED"),
	DenoiseFilter:  sToPtr("DISABLED"),
	FilterEnable:   sToPtr("AUTO"),
	FilterStrength: i64ToPtr(0),
	PsiControl:     sToPtr("USE_PSI"),
	TimecodeSource: sToPtr("EMBEDDED"),
	VideoSelector: &mediaconvert.VideoSelector{
		ColorSpace: sToPtr("FOLLOW")}}
var outputs = []*mediaconvert.Output{
	&mediaconvert.Output{
		Preset:       sToPtr("System-Avc_16x9_720p_29_97fps_3500kbps"),
		NameModifier: sToPtr("_high")},
	&mediaconvert.Output{
		Preset:       sToPtr("System-Avc_16x9_360p_29_97fps_1200kbps"),
		NameModifier: sToPtr("_mid")},
	&mediaconvert.Output{
		Preset:       sToPtr("System-Avc_16x9_270p_14_99fps_400kbps"),
		NameModifier: sToPtr("_low")}}
var hlsGroupSettings = mediaconvert.HlsGroupSettings{
	CaptionLanguageSetting: sToPtr("OMIT"),
	ClientCache:            sToPtr("ENABLED"),
	CodecSpecification:     sToPtr("RFC_4281"),
	DirectoryStructure:     sToPtr("SINGLE_DIRECTORY"),
	ManifestCompression:    sToPtr("NONE"),
	ManifestDurationFormat: sToPtr("INTEGER"),
	MinSegmentLength:       i64ToPtr(0),
	OutputSelection:        sToPtr("MANIFESTS_AND_SEGMENTS"),
	ProgramDateTime:        sToPtr("EXCLUDE"),
	ProgramDateTimePeriod:  i64ToPtr(600),
	SegmentControl:         sToPtr("SEGMENTED_FILES"),
	SegmentLength:          i64ToPtr(10),
	StreamInfResolution:    sToPtr("INCLUDE"),
	TimedMetadataId3Frame:  sToPtr("PRIV"),
	TimedMetadataId3Period: i64ToPtr(10)}
var outputGroupSettings = mediaconvert.OutputGroupSettings{
	Type: sToPtr("HLS_GROUP_SETTINGS")}
var outputGroup = mediaconvert.OutputGroup{
	CustomName: sToPtr("hls"),
	Name:       sToPtr("Apple HLS"),
	Outputs:    outputs}
var createJobInput = &mediaconvert.CreateJobInput{
	Queue: &queue,
	Role:  &role}

// convertHandler is a function to handle event
func convertHandler(ctx context.Context, event events.S3Event) error {

	// Get event data
	eventJSON, err := json.Marshal(event)
	if err != nil {
		log.Print(err)
		return nil
	}
	log.Print(string(eventJSON))

	// Get variables from event data
	params := event.Records[0].S3
	bucket := params.Bucket.Name
	key := params.Object.Key
	size := params.Object.Size

	// Handle only file size of which is not zero
	if size == 0 {
		log.Print("Not media file.")
		return nil
	}

	// Get customer endpoint of MediaConvert
	endpoints, err := svc.DescribeEndpoints(&mediaconvert.DescribeEndpointsInput{})
	if err != nil {
		log.Print(err)
		return err
	}
	endpointURL := endpoints.Endpoints[0].Url
	client := mediaconvert.New(sess, config.WithEndpoint(*endpointURL))

	// Set variables to setting struct
	input.SetFileInput(fmt.Sprintf("s3://%s/%s", bucket, key))
	hgs := hlsGroupSettings.SetDestination(fmt.Sprintf("s3://%s/%s", destBucket, key))
	ogs := outputGroupSettings.SetHlsGroupSettings(hgs)
	og := outputGroup.SetOutputGroupSettings(ogs)
	jobSettings := settings.
		SetInputs([]*mediaconvert.Input{input}).
		SetOutputGroups([]*mediaconvert.OutputGroup{og})
	createJobInput = &mediaconvert.CreateJobInput{
		Queue:    &queue,
		Role:     &role,
		Settings: jobSettings.SetInputs([]*mediaconvert.Input{input}).SetOutputGroups([]*mediaconvert.OutputGroup{&outputGroup})}

	// Call CreateJob
	result, err := client.CreateJob(createJobInput)
	if err != nil {
		log.Print(err)
		return err
	}
	log.Print(strings.Replace(result.GoString(), "\n", "", -1))
	return nil
}

// main is a function to call convertHandler
func main() {
	lambda.Start(convertHandler)
}
