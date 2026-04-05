package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var (
	port          string
	model         string
	streaming     bool
	verbose       bool
	storageFolder string
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

var rootCmd = &cobra.Command{
	Use:   "go-copilotkit",
	Short: "CopilotKit runtime server",
	Run: func(cmd *cobra.Command, args []string) {
		_ = godotenv.Load()

		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			log.Fatal("GEMINI_API_KEY environment variable required")
		}

		// Model priority: flag > GEMINI_MODEL > MODEL
		if model == "" {
			model = os.Getenv("GEMINI_MODEL")
			if model == "" {
				model = os.Getenv("MODEL")
			}
		}

		// Streaming from env if not set by flag
		if !cmd.Flags().Changed("streaming") {
			streaming = os.Getenv("GEMINI_STREAMING") == "true"
		}

		// Verbose from env if not set by flag
		if !cmd.Flags().Changed("verbose") {
			verbose = os.Getenv("VERBOSE") == "true"
		}

		// Port from env if not set by flag
		if !cmd.Flags().Changed("port") && os.Getenv("PORT") != "" {
			port = os.Getenv("PORT")
		}

		// Storage folder from env if not set by flag
		if !cmd.Flags().Changed("storage-folder") && os.Getenv("STORAGE_FOLDER") != "" {
			storageFolder = os.Getenv("STORAGE_FOLDER")
		}

		storage, err := NewFileStorage(storageFolder)
		if err != nil {
			log.Fatal(err)
		}

		agent, err := NewGeminiAgent(apiKey, model, verbose, streaming)
		if err != nil {
			log.Fatal(err)
		}

		protocol := NewProtocol(agent, storage)
		http.Handle("/copilotkit", loggingMiddleware(http.HandlerFunc(protocol.Handler)))

		log.Printf("CopilotKit server running on :%s (model=%s, streaming=%v, verbose=%v, storage=%s)", port, model, streaming, verbose, storageFolder)
		log.Fatal(http.ListenAndServe(":"+port, nil))
	},
}

func init() {
	rootCmd.Flags().StringVarP(&port, "port", "p", "4000", "Server port")
	rootCmd.Flags().StringVarP(&model, "model", "m", "", "Gemini model name")
	rootCmd.Flags().BoolVarP(&streaming, "streaming", "s", false, "Enable streaming mode")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose logging")
	rootCmd.Flags().StringVar(&storageFolder, "storage-folder", "./storage", "Folder to store conversation threads")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
