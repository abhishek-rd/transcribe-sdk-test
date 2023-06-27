package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/transcribestreaming"
	"github.com/hajimehoshi/oto"
	"github.com/hajimehoshi/oto/audio"
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

	// Create an audio context for streaming transcription
	audioContext, err := oto.NewContext(decoder.SampleRate(), 1, 2, 8192)
	if err != nil {
		log.Fatal("Failed to create audio context:", err)
	}
	defer audioContext.Close()

	// Start the transcription streaming job
	transcriptionStream, err := transcribeClient.StartStreamTranscriptionWithContext(context.Background(), &transcribestreaming.StartStreamTranscriptionInput{
		LanguageCode:       aws.String(languageCode),
		MediaEncoding:      aws.String(transcribestreaming.MediaEncodingOggOpus),
		MediaSampleRateHertz: aws.Int64(int64(decoder.SampleRate())),
	})
	if err != nil {
		log.Fatal("Failed to start transcription streaming:", err)
	}

	// Create an audio player for streaming audio
	audioPlayer := audioContext.NewPlayer()

	// Stream the audio data and send it for transcription
	bufferSize := 8192
	buffer := make([]int16, bufferSize)
	for {
		n, err := decoder.Read(buffer)
		if err != nil {
			break
		}

		// Play the audio data
		audioPlayer.Write(buffer[:n*2])

		// Create an audio event for the PCM data
		audioEvent := &transcribestreaming.AudioEvent{
			AudioChunk: int16ArrayToByte(buffer[:n]),
		}

		// Send the audio event to the transcription service
		_, err = transcriptionStream.AudioStream.Write(audioEvent.AudioChunk)
		if err != nil {
			log.Fatal("Failed to send audio chunk:", err)
		}
	}

	// Wait for the transcription job to complete
	err = transcriptionStream.AudioStream.Close()
	if err != nil {
		log.Fatal("Failed to close audio stream:", err)
	}

	// Receive and print the transcription results
	for {
		resp, err := transcriptionStream.Recv()
		if err != nil {
			log.Fatal("Failed to receive transcription response:", err)
		}

		for _, result := range resp.Results {
			for _, alt := range result.Alternatives {
				fmt.Println(*alt.Transcript)
			}
		}

		if resp.StreamingStatusCode == transcribestreaming.StreamingStatusCompleted {
			break
		}
	}

	// Stop playing audio
	audioPlayer.Close()
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
