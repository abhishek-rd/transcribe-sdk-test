// ERROR
//PS D:\Development\Transcribe\test> go run main.go
//main.go:6:2: no required module provides package github.com/aws/aws-sdk-go-v2/service/transcribestreamingservice; to add it:
//        go get github.com/aws/aws-sdk-go-v2/service/transcribestreamingservice
//PS D:\Development\Transcribe\test> go get github.com/aws/aws-sdk-go-v2/service/transcribestreamingservice
//go: module github.com/aws/aws-sdk-go-v2@upgrade found (v1.18.1), but does not contain package github.com/aws/aws-sdk-go-v2/service/transcribestreamingservice


package main

import (
    "fmt"
    "io"
	"github.com/aws/aws-sdk-go-v2/service/transcribestreamingservice"

)

func main() {
    // Create a transcribe streaming client
    client := transcribestreamingservice.NewClient()

    // Get the file to transcribe
    file, err := os.Open("audio.ogg")
    if err != nil {
        fmt.Println(err)
        return
    }

    // Split the file into chunks
    chunkSize := 1024 * 1024 // 1 MB
    var chunks [][]byte
    for {
        bytes := make([]byte, chunkSize)
        n, err := file.Read(bytes)
        if err == io.EOF {
            break
        } else if err != nil {
            fmt.Println(err)
            return
        }

        chunks = append(chunks, bytes)
    }

    // Transcribe the chunks
    for _, chunk := range chunks {
        // Create an audio event
        audioEvent := transcribestreamingservice.AudioEvent{
            AudioChunk: chunk,
        }

        // Send the audio event to the transcribe streaming client
        client.SendAudioEvent(audioEvent)
    }

    // Wait for the transcription to finish
    client.WaitUntilTranscriptionComplete()

    // Print the transcript
    fmt.Println(client.Transcript())
}
