package main

import (
	"encoding/json"
	"fmt"
	"github.com/Rhymen/go-whatsapp"
	"github.com/skip2/go-qrcode"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type WhatsappModel struct {
	whatsappConnect *whatsapp.Conn
	session         whatsapp.Session
}

func main() {
	wac, err := whatsapp.NewConn(100 * time.Second)

	if err != nil {
		log.Panic(err)
	}

	wac.SetClientVersion(3, 2123, 7)

	WhatsappModel := WhatsappModel{
		whatsappConnect: wac,
	}

	WhatsappModel.handleRequests()
}

func (WhatsappModel WhatsappModel) sendMessage(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	reqBody, _ := ioutil.ReadAll(r.Body)

	data := struct {
		Phone   string `json:"phone"`
		Message string `json:"message"`
	}{}

	err := json.Unmarshal(reqBody, &data)

	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	if data.Message == "" || data.Phone == "" {
		fmt.Fprintf(w, "Format Not Required")
		return
	}

	if !WhatsappModel.whatsappConnect.GetLoggedIn() {
		newSession, err := WhatsappModel.whatsappConnect.RestoreWithSession(WhatsappModel.readLastSession())
		if err != nil {
			fmt.Println("Whatsapp Disconnect")
		}
		WhatsappModel.session = newSession
	}

	text := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: data.Phone + "@s.whatsapp.net",
		},
		Text: data.Message,
	}

	_, err = WhatsappModel.whatsappConnect.Send(text)

	if err != nil {
		fmt.Fprintf(w, err.Error())
		return
	}

	jsonInBytes, err := json.Marshal(data)

	w.Header().Set("Content-Type", "application/json")

	w.Write(jsonInBytes)
}

func (WhatsappModel WhatsappModel) readLastSession() whatsapp.Session {
	jsonFile, err := os.Open("session.json")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("Successfully Opened session.json")

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	session := whatsapp.Session{}

	json.Unmarshal([]byte(byteValue), &session)

	return session
}

func (WhatsappModel WhatsappModel) loginWhatsapp(w http.ResponseWriter, r *http.Request) {

	wac, err := whatsapp.NewConn(100 * time.Second)

	if err != nil {
		log.Panic(err)
	}

	wac.SetClientVersion(3, 2123, 7)

	WhatsappModel.whatsappConnect = wac

	qr := make(chan string)

	go func() {
		err := qrcode.WriteFile(<- qr, qrcode.Medium, 256, "scan_qr_ini.png")

		if err != nil {
			fmt.Println(err.Error())
		}
	}()

	session, err := WhatsappModel.whatsappConnect.Login(qr)

	file, _ := json.MarshalIndent(session, "", " ")

	_ = ioutil.WriteFile("session.json", file, 0644)

	if err != nil {

		message := struct {
			Message string `json:"message"`
		}{
			Message: err.Error(),
		}

		jsonInBytes, err := json.Marshal(message)

		if err != nil {
			fmt.Println(err)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonInBytes)

		return
	}

	message := struct {
		Message string `json:"message"`
	}{
		Message: "success",
	}

	jsonInBytes, err := json.Marshal(message)

	w.Header().Set("Content-Type", "application/json")

	w.Write(jsonInBytes)
}

func (WhatsappModel WhatsappModel) getQr(w http.ResponseWriter, r *http.Request) {
	jsonFile, err := os.Open("qrcode.json")

	if err != nil {
		fmt.Println(err)
	}

	defer jsonFile.Close()

	byteValue, _ := ioutil.ReadAll(jsonFile)

	qrCode := struct {
		QrCode string `json:"qr_code"`
	}{}

	json.Unmarshal([]byte(byteValue), &qrCode)

	jsonInBytes, err := json.Marshal(qrCode)

	w.Header().Set("Content-Type", "application/json")

	w.Write(jsonInBytes)
}

func (WhatsappModel WhatsappModel) handleRequests() {
	http.HandleFunc("/login", WhatsappModel.loginWhatsapp)
	http.HandleFunc("/send", WhatsappModel.sendMessage)
	http.HandleFunc("/getQr", WhatsappModel.getQr)
	log.Fatal(http.ListenAndServe(":8000", nil))
}
