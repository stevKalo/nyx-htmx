package main

import (
	"bufio"
	"fmt"
	"html"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	goaway "github.com/TwiN/go-away"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func welcomeHandler(w http.ResponseWriter, req *http.Request) {
	err := NyxMessage("Hello, my name is Nyx. People like to tell me messages, would you like to hear one?").Render(req.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering welcome message", http.StatusInternalServerError)
		return
    }

	err = YesNo("listen").Render(req.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering choice box", http.StatusInternalServerError)
		return
    }
}

func listenHandler(w http.ResponseWriter, req *http.Request) {

	answer := req.URL.Query().Get("answer")
	
	switch answer {
	case "yes": 
		message, err := getMessage()
		if err != nil {
			http.Error(w, "Error getting message", http.StatusInternalServerError)
			return
		}

		err = voidMessage(message).Render(req.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering message", http.StatusInternalServerError)
			return
		}
		return
	case "no":
		err := NyxMessage("Very well, goodbye.").Render(req.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering message", http.StatusInternalServerError)
			return
		}
		return
	default:
		http.Error(w, "Invalid Answer!", http.StatusBadRequest)
		return
	}
}

func speakHandler(w http.ResponseWriter, req *http.Request){

}

func submitHandler(w http.ResponseWriter, req *http.Request) {

	err := req.ParseForm()
    if err != nil {
        http.Error(w, "Error reading body", http.StatusBadRequest)
        return
    }

	if len(req.FormValue("userInput")) > 500 {
		http.Error(w, "Message is too long!", http.StatusBadRequest)
		return
	}

	message := html.EscapeString(req.FormValue("userInput"))

	if goaway.IsProfane(message) {
		http.Error(w, "Message is profane!", http.StatusAccepted)
		err = NyxMessage("Message is profane!").Render(req.Context(), w)
		if err != nil {
			http.Error(w, "Error rendering NyxMessage", http.StatusInternalServerError)
			return
    	}
		return
	}

	err = saveMessage(message)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed saving file to S3: %v", err.Error()), http.StatusBadRequest)
		return
	}

    err = UserMessage(message).Render(req.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering UserMessage", http.StatusInternalServerError)
		return
    }

    err = NyxMessage("Nice message!").Render(req.Context(), w)
	if err != nil {
		http.Error(w, "Error rendering NyxMessage", http.StatusInternalServerError)
		return
    }
}

func saveMessage(newMessage string) error {
	sess, err := session.NewSession(&aws.Config{
        Region: aws.String("us-east-2"), // Set your desired region
    })
	if err != nil {
        return fmt.Errorf("failed to create AWS session: %v", err.Error())
    }

	err = downloadFile(sess)
	if err != nil {
		return fmt.Errorf("failed to download file: %v", err.Error())
	}

	messages, err := readMessages()
	if err != nil {
		return fmt.Errorf("failed to read messages: %v", err.Error())
	}

	if isDuplicate(messages, newMessage) {
		fmt.Println("Message is a duplicate!")
		return nil
	}

	messages = append(messages, newMessage)

	err = writeMessages(messages)
	if err != nil {
		return fmt.Errorf("failed to write messages: %v", err.Error())
	}

	err = uploadFile(sess)
	if err != nil {
		return fmt.Errorf("failed to upload file: %v", err.Error())
	}
	// TODO: add some kind of lock for multi-server

	return nil
}

func readMessages() ([]string, error) {
	file, err := os.Open("messages.txt")
	if err != nil {
		return nil, fmt.Errorf("error opening messages.txt: %v", err.Error())
	}

	defer file.Close()

	var messages []string

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			messages = append(messages, line)
		}
	} 

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan file: %v", err)
	}

	return messages, nil
}

func getMessage() (string, error) {
	sess, err := session.NewSession(&aws.Config{
        Region: aws.String("us-east-2"), // Set your desired region
    })
	if err != nil {
        return "", fmt.Errorf("failed to create AWS session: %v", err.Error())
    }

	err = downloadFile(sess)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %v", err.Error())
	}

	messages, err := readMessages()
	if err != nil {
		return "", fmt.Errorf("failed to read messages: %v", err.Error())
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
    randomIndex := r.Intn(len(messages))

	return messages[randomIndex], nil
}

func writeMessages(messages []string) error {
	file, err := os.Create("messages.txt")
	if err != nil {
		return fmt.Errorf("error opening messages.txt: %v", err.Error())
	}

	defer file.Close()

    writer := bufio.NewWriter(file)
    defer writer.Flush()

    for _, message := range messages {
        _, err := writer.WriteString(message + "\n")
        if err != nil {
            return err
        }
    }

	// TODO: delete messages file locally

    return nil
}

func isDuplicate(messages []string, newMessage string) bool {
    for _, msg := range messages {
        if msg == newMessage {
            return true
        }
    }
    return false
}