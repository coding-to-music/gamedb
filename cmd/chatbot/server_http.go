package main

import (
	"compress/flate"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot"
	"github.com/gamedb/gamedb/pkg/chatbot/interactions"
	"github.com/gamedb/gamedb/pkg/config"
	"github.com/gamedb/gamedb/pkg/log"
	"github.com/gamedb/gamedb/pkg/memcache"
	"github.com/gamedb/gamedb/pkg/middleware"
	"github.com/gamedb/gamedb/pkg/mysql"
	"github.com/go-chi/chi"
	chiMiddleware "github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

func slashCommandServer() error {

	r := chi.NewRouter()
	r.Use(chiMiddleware.RedirectSlashes)
	r.Use(chiMiddleware.NewCompressor(flate.DefaultCompression, "text/plain", "application/json").Handler)
	r.Use(middleware.RealIP)

	r.Post("/", discordHandler)
	r.Get("/health-check", healthCheckHandler)

	r.NotFound(notFoundHandler)

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
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	signature := r.Header.Get("X-Signature-Ed25519")
	timestamp := r.Header.Get("X-Signature-Timestamp")

	// Verify against signature
	valid, err := verifyRequest(signature, timestamp+string(body), config.C.DiscordOChatBotPublKey)
	if err != nil {
		log.ErrS(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	if !valid {
		http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusInternalServerError)
		return
	}

	event := interactions.Event{}
	err = json.Unmarshal(body, &event)
	if err != nil {
		log.ErrS(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Check for pings
	if event.Type == 1 {

		response := interactions.Response{
			Type: interactions.ResponseTypePong,
		}

		b, _ := json.Marshal(response)
		_, _ = w.Write(b)
		return
	}

	// Get command
	command, ok := chatbot.CommandCache[event.Data.Name]
	if !ok {
		http.Error(w, "Command ID not found in register", http.StatusNotFound)
		return
	}

	// Save stats
	defer saveToDB(
		command,
		event.Arguments(),
		true,
		event.GuildID,
		event.ChannelID,
		event.Member.User.ID,
		event.Member.User.Username,
		event.Member.User.Avatar,
	)

	discordSession, err := getSession()
	if err != nil {
		log.ErrS(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Typing notification
	err = discordSession.ChannelTyping(event.ChannelID)
	discordError(err)

	// Get user settings
	code := steamapi.ProductCCUS
	if command.PerProdCode() {
		settings, err := mysql.GetChatBotSettings(event.Member.User.ID)
		if err != nil {
			log.ErrS(err)
		}
		code = settings.ProductCode
	}

	cacheItem := memcache.ItemChatBotRequestSlash(command.ID(), event.Arguments(), code)

	// Check in cache first
	if !command.DisableCache() && !config.IsLocal() {

		var response interactions.Response
		err = memcache.GetInterface(cacheItem.Key, &response)
		if err == nil {

			b, _ := json.Marshal(response)
			_, _ = w.Write(b)
			return
		}
	}

	// Rate limit
	if !limits.GetLimiter(event.Member.User.ID).Allow() {
		log.Warn("over chatbot rate limit", zap.String("author", event.Member.User.ID), zap.String("msg", event.ArgumentsString()))
		http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		return
	}

	out, err := command.Output(event.Member.User.ID, code, event.Arguments())
	if err != nil {
		log.ErrS(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := interactions.Response{
		Type: interactions.ResponseTypeMessageWithSource,
		Data: interactions.ResponseData{
			TTS:     false,
			Content: out.Content,
			Embeds:  []*discordgo.MessageEmbed{out.Embed},
		},
	}

	// Save to cache
	err = memcache.SetInterface(cacheItem.Key, response, cacheItem.Expiration)
	if err != nil {
		log.Err("Saving to memcache", zap.Error(err), zap.String("msg", event.ArgumentsString()))
	}

	// Respond
	b, _ := json.Marshal(response)
	_, err = w.Write(b)
	if err != nil {
		log.ErrS(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
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

func notFoundHandler(w http.ResponseWriter, _ *http.Request) {
	http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
}
