package main

import (
	"compress/flate"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/Jleagle/steam-go/steamapi"
	"github.com/bwmarrin/discordgo"
	"github.com/gamedb/gamedb/pkg/chatbot"
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
	r.Use(chiMiddleware.NewCompressor(flate.DefaultCompression, "text/plain", "application/json").Handler)
	r.Use(middleware.RealIP)

	r.Get("/", healthCheckHandler)
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

	body, err := io.ReadAll(r.Body)
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

	interaction := &discordgo.Interaction{}
	err = json.Unmarshal(body, interaction)
	if err != nil {
		log.ErrS(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	// Check for pings
	if interaction.Type == discordgo.InteractionPing {

		response := discordgo.InteractionResponse{
			Type: discordgo.InteractionResponsePong,
		}

		b, _ := json.Marshal(response)
		_, _ = w.Write(b)
		return
	}

	// Get command
	command, ok := chatbot.CommandCache[interaction.Data.Name]
	if !ok {
		http.Error(w, "Command ID not found in register", http.StatusNotFound)
		return
	}

	// Save stats
	defer saveToDB(command, true, argumentsString(interaction), interaction.GuildID, interaction.ChannelID, interaction.Member.User)

	// Typing notification
	err = discordSession.ChannelTyping(interaction.ChannelID)
	discordError(err)

	// Get user settings
	code := steamapi.ProductCCUS
	if command.PerProdCode() {
		settings, err := mysql.GetChatBotSettings(interaction.Member.User.ID)
		if err != nil {
			log.ErrS(err)
		}
		code = settings.ProductCode
	}

	cacheItem := memcache.ItemChatBotRequestSlash(command.ID(), arguments(interaction), code)

	// Check in cache first
	if !command.DisableCache() && !config.IsLocal() {

		var response discordgo.InteractionResponse
		err = memcache.GetInterface(cacheItem.Key, &response)
		if err == nil {

			b, _ := json.Marshal(response)
			_, _ = w.Write(b)
			return
		}
	}

	// Rate limit
	if !limits.GetLimiter(interaction.Member.User.ID).Allow() {
		log.Warn("over chatbot rate limit", zap.String("author", interaction.Member.User.ID), zap.String("msg", argumentsString(interaction)))
		http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
		return
	}

	out, err := command.Output(interaction.Member.User.ID, code, arguments(interaction))
	if err != nil {
		log.ErrS(err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	response := discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionApplicationCommandResponseData{
			Content: out.Content,
		},
	}

	if out.Embed != nil {
		response.Data.Embeds = []*discordgo.MessageEmbed{out.Embed}
	}

	// Save to cache
	err = memcache.SetInterface(cacheItem.Key, response, cacheItem.Expiration)
	if err != nil {
		log.Err("Saving to memcache", zap.Error(err), zap.String("msg", argumentsString(interaction)))
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

// todo, use the function in discordgo
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
