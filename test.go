package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/transcribe"
	"github.com/aws/aws-sdk-go/service/transcribe/transcribestreaming"
)

func main() {
	// Replace "your-file-path" with the path to your audio file
	filePath := "your-file-path"

	// Create an AWS session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	// Create a Transcribe Streaming client
	streamingClient := transcribestreaming.New(sess)

	// Start a stream transcription
	startStreamTranscriptionInput := &transcribestreaming.StartStreamTranscriptionInput{
		LanguageCode: transcribe.LanguageCodeEnUS, // Replace with the desired language code
		MediaSampleRateHertz: aws.Int64(44100),     // Replace with the actual sample rate of your audio file
		MediaEncoding:       aws.String(transcribestreaming.MediaEncodingPcm),
	}
	stream, err := streamingClient.StartStreamTranscription(startStreamTranscriptionInput)
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
		audioEvent := &transcribestreaming.AudioEvent{
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

	// Get the transcription results
	getTranscriptResultStreamInput := &transcribestreaming.GetTranscriptResultStreamInput{
		RequestId: stream.RequestId,
	}
	results, err := streamingClient.GetTranscriptResultStream(getTranscriptResultStreamInput)
	if err != nil {
		fmt.Println("Failed to get transcript result stream:", err)
		return
	}

	// Print the transcript
	for result := range results {
		if result.IsPartial() {
			// Handle partial result
			fmt.Println("Partial transcript:", *result.Transcript.Results[0].Alternatives[0].Transcript)
		} else {
			// Handle final result
			fmt.Println("Final transcript:", *result.Transcript.Results[0].Alternatives[0].Transcript)
		}
	}
}
