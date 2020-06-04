package main

import (
	"bytes"
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

type timeSlot struct {
	Date         string `json:"date"`
	Start        string `json:"start"`
	End          string `json:"end"`
	Availability string `json:"availability"`
}

type sendMessageRequestBody struct {
	ChatID  string
	Message string
}

func main() {
	sentTimeSlot := make(map[string]bool)
	for {
		text := ""
		timeSlots, err := getAvaliableTimeSlots()
		if err == nil && len(timeSlots) > 0 {
			for _, slot := range timeSlots {
				key := slot.Date + slot.Start
				_, sent := sentTimeSlot[key]

				if !sent {
					text += "Date: " + slot.Date + " Time: " + slot.Start + "\n"
					sentTimeSlot[key] = true
				}
			}
			sendNotification(text)
		}

		time.Sleep(time.Second * 30)
	}
}

func getAvaliableTimeSlots() ([]timeSlot, error) {
	result := []timeSlot{}
	guid := randomString(36)
	res, err := http.Get("https://sushipass.sushiro.com.hk/api/1.1/info/reservationtimeslots?storeid=" + os.Getenv("STORE_ID") + "&numpersons=" + os.Getenv("NUM_PERSONS") + "&guid=" + guid)
	if err != nil {
		log.Println(err)
		return result, err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Printf("status code error: %d %s", res.StatusCode, res.Status)
		return result, err
	}

	var timeSlotResponse []timeSlot
	err = json.NewDecoder(res.Body).Decode(&timeSlotResponse)
	if err != nil {
		log.Println(err)
		return result, err
	}

	for _, slot := range timeSlotResponse {
		start, _ := strconv.Atoi(slot.Start)
		startTimeConfig, _ := strconv.Atoi(os.Getenv("START_TIME"))

		if slot.Availability != "UNAVAILABLE" && start > startTimeConfig {
			result = append(result, slot)
		}
	}

	return result, nil
}

func sendNotification(text string) {
	requestBody, _ := json.Marshal(map[string]string{
		"chat_id": os.Getenv("CHAT_ID"),
		"text":    text,
	})

	http.Post("https://api.telegram.org/bot"+os.Getenv("TELEGRAM_TOKEN")+"/sendMessage", "application/json", bytes.NewBuffer(requestBody))
}

func randomString(n int) string {
	var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}
