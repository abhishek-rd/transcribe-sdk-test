package main

import (
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/transcribestreamingservice"
)

func main() {
	// Replace "your-file-path" with the path to your OGG file
	filePath := "your-file-path"

	// Create an AWS session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create a Transcribe Streaming client
	client := transcribestreamingservice.New(sess)

	// Create a stream for transcription
	stream, err := client.StartStreamTranscription(&transcribestreamingservice.StartStreamTranscriptionInput{
		LanguageCode:        aws.String("en-US"), // Replace with the desired language code
		MediaSampleRateHertz: aws.Int64(16000),   // Replace with the actual sample rate of your audio file
		MediaEncoding:       aws.String("ogg-opus"),
	})
	if err != nil {
		fmt.Println("Failed to start stream transcription:", err)
		return
	}

	// Open the audio file
	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Failed to open file:", err)
		return
	}
	defer file.Close()

	// Chunk size in bytes
	chunkSize := 1024 * 16

	// Buffer to hold audio data
	buffer := make([]byte, chunkSize)

	// Read and stream the audio file in chunks
	for {
		// Read a chunk of audio data
		bytesRead, err := file.Read(buffer)
		if err != nil {
			fmt.Println("Failed to read from file:", err)
			return
		}

		// Check if reached end of file
		if bytesRead == 0 {
			break
		}

		// Create an audio event
		audioEvent := &transcribestreamingservice.AudioEvent{
			AudioChunk: buffer[:bytesRead],
		}

		// Send the audio event to the transcription stream
		err = stream.Send(audioEvent)
		if err != nil {
			fmt.Println("Failed to send audio event:", err)
			return
		}
	}

	// Close the stream
	_, err = stream.Close()
	if err != nil {
		fmt.Println("Failed to close stream:", err)
		return
	}

	// Process the event stream for transcription results
	for {
		event, err := stream.Recv()
		if err != nil {
			fmt.Println("Failed to receive event:", err)
			return
		}

		switch event.GetEventType() {
		case transcribestreamingservice.EventTypeTranscript:
			transcriptEvent := event.GetTranscriptEvent()
			if transcriptEvent != nil {
				results := transcriptEvent.GetTranscript().GetResults()
				for _, result := range results {
					if result.IsPartial() {
						// Handle partial result
						fmt.Println("Partial transcript:", *result.GetAlternatives()[0].GetTranscript())
					} else {
						// Handle final result
						fmt.Println("Final transcript:", *result.GetAlternatives()[0].GetTranscript())
					}
				}
			}
		}
	}
}
