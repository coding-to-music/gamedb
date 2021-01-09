package main

import (
	"compress/flate"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/middleware"
	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
)

func slashCommandServer() error {

	r := chi.NewRouter()
	r.Use(chiMiddleware.RedirectSlashes)
	r.Use(chiMiddleware.NewCompressor(flate.DefaultCompression, "text/html", "text/css", "text/javascript", "application/json", "application/javascript").Handler)
	r.Use(middleware.RealIP)

	r.Post("/", discordHandler)
	r.Get("/health-check", healthCheckHandler)

	for _, c := range chatbot.CommandRegister {

		r.Get("/"+c.ID(), func(w http.ResponseWriter, r *http.Request) {

			_, err := w.Write([]byte("success"))
			if err != nil {
				log.ErrS(err)
			}
		})
	}

	r.NotFound(errorHandler)

	s := &http.Server{
		Addr:              "0.0.0.0:" + config.C.ChatbotPort,
		Handler:           r,
		ReadTimeout:       2 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}

	log.Info("Starting Chatbot webserver on http://" + s.Addr + "/")

	go func() {
		err := s.ListenAndServe() // Blocks
		if err != nil {
			log.ErrS(err)
		}
	}()

	return nil
}

func discordHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.ErrS(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	signature := r.Header.Get("X-Signature-Ed25519")
	timestamp := r.Header.Get("X-Signature-Timestamp")

	// Verify against signature
	valid, err := verifyRequest(signature, timestamp+string(body), config.C.DiscordOChatBotPublKey)
	if err != nil {
		log.ErrS(err)
	}
	if !valid {
		http.Error(w, http.StatusText(401), 401)
		return
	}

	event := interactions.Event{}
	err = json.Unmarshal(body, &event)
	if err != nil {
		log.ErrS(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// Check for pings
	if event.Type == 1 {

		response := interactions.Response{
			Type: interactions.ResponseTypePong,
		}

		b, err := json.Marshal(response)
		if err != nil {
			log.ErrS(err)
			http.Error(w, http.StatusText(500), 500)
			return
		}

		_, _ = w.Write(b)
		return
	}

	command, ok := chatbot.CommandCache[event.Data.Name]
	if !ok {
		http.Error(w, "Command ID not found in register", 404)
		return
	}

	// Convert to old style input
	var oldStyle = []string{"." + command.LegacyPrefix()}
	for _, v := range event.Data.Options {
		oldStyle = append(oldStyle, v.Value)
	}

	payload := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			Content: strings.Join(oldStyle, " "),
			Author: &discordgo.User{
				ID: event.Member.User.ID,
			},
		},
	}

	out, err := command.Output(payload, steamapi.ProductCCUS)
	if err != nil {
		log.ErrS(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	// Don't need footer on slash commands
	out.Embed.Footer = nil

	response := interactions.Response{
		Type: interactions.ResponseTypeMessageWithSource,
		Data: interactions.ResponseData{
			TTS:     false,
			Content: out.Content,
			Embeds:  []*discordgo.MessageEmbed{out.Embed},
		},
	}

	b, err := json.Marshal(response)
	if err != nil {
		log.ErrS(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}

	_, err = w.Write(b)
	if err != nil {
		log.ErrS(err)
		http.Error(w, http.StatusText(500), 500)
		return
	}
}

func verifyRequest(signature, hash, publicKey string) (bool, error) {

	decodedSignature, err := hex.DecodeString(signature)
	if err != nil {
		return false, err
	}

	decodedPublicKey, err := hex.DecodeString(publicKey)
	if err != nil {
		return false, err
	}

	return ed25519.Verify(decodedPublicKey, []byte(hash), decodedSignature), nil
}

func healthCheckHandler(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, http.StatusText(http.StatusOK), http.StatusOK)
}

func errorHandler(w http.ResponseWriter, _ *http.Request) {

	w.WriteHeader(404)

	_, err := w.Write([]byte("404"))
	if err != nil {
		log.ErrS(err)
	}
}
