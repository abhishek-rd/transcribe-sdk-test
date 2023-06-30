package main

import (
	"context"
	"fmt"
	"io/ioutil"
	//"os"

	speech "cloud.google.com/go/speech/apiv1"
	speechpb "google.golang.org/genproto/googleapis/cloud/speech/v1"
	"google.golang.org/api/option"
)

func main() {
	// Path to your JSON key file
	keyFilePath := "integral-glass-391419-702a3e04aee0.json"

	// Create a new SpeechClient with authentication using a JSON key file
	ctx := context.Background()
	client, err := speech.NewClient(ctx, option.WithCredentialsFile(keyFilePath))
	if err != nil {
		fmt.Println(err)
		return
	}

	// Read the OGG audio file.
	filePath := "tests_integration_assets_test.wav"
	audioData, err := ioutil.ReadFile(filePath)
	if err != nil {
		fmt.Println(err)
		return
	}
	//en := "pcm"
	// Create a new RecognitionConfig with the desired configuration.
	config := &speechpb.RecognitionConfig{
		Encoding:        speechpb.RecognitionConfig_LINEAR16, //speechpb.RecognitionConfig_OGG_OPUS,
		SampleRateHertz: 16000,
		LanguageCode:    "en-US",
	}

	// Create a new RecognitionAudio from the audio data.
	audio := &speechpb.RecognitionAudio{
		AudioSource: &speechpb.RecognitionAudio_Content{Content: audioData},
	}

	// Perform the speech recognition.
	resp, err := client.Recognize(ctx, &speechpb.RecognizeRequest{
		Config: config,
		Audio:  audio,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	// Print the transcription.
	for _, result := range resp.Results {
		fmt.Println(result.Alternatives[0].Transcript)
	}
}
