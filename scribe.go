package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/transcribestreaming"
	"github.com/hajimehoshi/oto"
	"github.com/youpy/go-oggvorbis"
)

func main() {
	// AWS Transcribe configuration
	region := "us-west-2"
	languageCode := "en-US"

	// Path to the Ogg audio file
	audioFilePath := "/path/to/your/audio_file.ogg"

	// Create a session using your AWS credentials
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Config: aws.Config{Region: aws.String(region)},
	}))

	// Create an AWS Transcribe Streaming service client
	transcribeClient := transcribestreaming.New(sess)

	// Open the Ogg audio file
	file, err := os.Open(audioFilePath)
	if err != nil {
		log.Fatal("Failed to open audio file:", err)
	}
	defer file.Close()

	// Decode the Ogg audio file
	decoder, err := oggvorbis.NewReader(file)
	if err != nil {
		log.Fatal("Failed to decode audio file:", err)
	}

	// Create an audio stream for streaming transcription
	audioStream := transcribestreaming.NewAudioEventStream()

	// Start the transcription streaming job
	startStreamTranscription(transcribeClient, audioStream, languageCode)

	// Stream the audio data and send it for transcription
	buffer := make([]int16, 1024)
	for {
		_, err := decoder.Read(buffer)
		if err != nil {
			break
		}

		// Convert the audio data to PCM format
		pcmData := int16ArrayToByte(buffer)

		// Create an audio event for the PCM data
		audioEvent := &transcribestreaming.AudioEvent{
			AudioChunk: pcmData,
		}

		// Send the audio event to the transcription service
		audioStream.Input <- audioEvent

		// Delay to simulate real-time streaming (adjust as needed)
		time.Sleep(100 * time.Millisecond)
	}

	// Close the audio stream
	audioStream.Close()

	// Wait for the transcription job to complete
	waitForTranscriptionJob(transcribeClient)
}

// Start the transcription streaming job
func startStreamTranscription(transcribeClient *transcribestreaming.TranscribeStreaming, audioStream *transcribestreaming.AudioEventStream, languageCode string) {
	input := &transcribestreaming.StartStreamTranscriptionInput{
		LanguageCode: aws.String(languageCode),
		MediaSampleRateHertz: aws.Int64(48000), // Sample rate of the audio file (adjust as needed)
		MediaEncoding:       aws.String(transcribestreaming.MediaEncodingOggOpus),
	}

	go func() {
		_, err := transcribeClient.StartStreamTranscription(input, audioStream)
		if err != nil {
			log.Fatal("Failed to start transcription streaming:", err)
		}
	}()
}

// Wait for the transcription job to complete
func waitForTranscriptionJob(transcribeClient *transcribestreaming.TranscribeStreaming) {
	log.Println("Waiting for transcription job to complete...")

	for {
		// Check the status of the transcription job
		// You can add logic here to handle interim results if desired
		output, err := transcribeClient.DescribeStreamTranscriptionJob(&transcribestreaming.DescribeStreamTranscriptionJobInput{})
		if err != nil {
			log.Fatal("Failed to describe transcription job:", err)
		}

		jobStatus := aws.StringValue(output.StreamTranscriptionJob.StreamTranscriptionJobStatus)
		if jobStatus == transcribestreaming.StreamTranscriptionStatusCompleted {
			log.Println("Transcription job completed.")
			break
		} else if jobStatus == transcribestreaming.StreamTranscriptionStatusFailed {
			log.Fatal("Transcription job failed.")
		}

		// Delay before checking the job status again
		time.Sleep(5 * time.Second)
	}
}

// Convert an int16 array to a byte array
func int16ArrayToByte(data []int16) []byte {
	byteData := make([]byte, len(data)*2)
	for i := 0; i < len(data); i++ {
		byteData[i*2] = byte(data[i])
		byteData[i*2+1] = byte(data[i] >> 8)
	}
	return byteData
}
