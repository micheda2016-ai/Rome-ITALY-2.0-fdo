package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
)

var (
	Token            = os.Getenv("TOKEN")
	LogChannelID     = "ID_CANALE_LOG"      // Inserisci l'ID del canale log
	TicketCategoryID = "ID_CATEGORIA_TICKET" // Inserisci l'ID della categoria ticket
	TLSCertFile      = "cert.pem"           // Percorso del certificato SSL/TLS
	TLSKeyFile       = "key.pem"            // Percorso della chiave privata SSL/TLS
)

func main() {
	sess, err := discordgo.New("Bot " + Token)
	if err != nil {
		log.Fatalf("Errore nella creazione della sessione: %v", err)
	}

	// Intenti necessari per bottoni, modali e gestione membri
	sess.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsGuildMembers |
		discordgo.IntentsMessageContent

	sess.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		switch i.Type {
		case discordgo.InteractionApplicationCommand:
			handleSlashCommands(s, i)
		case discordgo.InteractionModalSubmit:
			handleModalSubmits(s, i)
		case discordgo.InteractionMessageComponent:
			handleComponents(s, i)
		}
	})

	sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Bot online come %s#%s", s.State.User.Username, s.State.User.Discriminator)
	})

	err = sess.Open()
	if err != nil {
		log.Fatalf("Errore durante l'apertura della connessione: %v", err)
	}
	defer sess.Close()

	// Configurazione e avvio del server HTTPS per il KeePassLive
	go startHTTPServer()

	// Registrazione comandi slash
	registerCommands(sess)

	fmt.Println("Bot avviato con successo. Premi CTRL-C per uscire.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

// Configurazione del server HTTPS
func startHTTPServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("HTTPS Server Active"))
	})

	server := &http.Server{
		Addr:    ":443",
		Handler: mux,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS12,
		},
	}

	log.Println("Avvio del server HTTPS sulla porta 443...")
	if err := server.ListenAndServeTLS(TLSCertFile, TLSKeyFile); err != nil {
		log.Printf("Nota sul server HTTPS: %v", err)
	}
}

func registerCommands(s *discordgo.Session) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "patente",
			Description: "Apri il modulo pop-up per richiedere la patente",
		},
		{
			Name:        "turno",
			Description: "Gestisci il tuo turno (Avvia/Termina)",
		},
		{
			Name:        "promuovi",
			Description: "Promuovi un membro dello staff",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "utente",
					Description: "Utente da promuovere",
					Required:    true,
				},
			},
		},
		{
			Name:        "retrocedi",
			Description: "Retrocedi un membro dello staff",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionUser,
					Name:        "utente",
					Description: "Utente da retrocedere",
					Required:    true,
				},
			},
		},
	}

	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			log.Printf("Errore nella creazione del comando %s: %v", cmd.Name, err)
		}
	}
}

func handleSlashCommands(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ApplicationCommandData()

	switch data.Name {
	case "patente":
		err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseModal,
			Data: &discordgo.InteractionResponseData{
				CustomID: "patente_modal",
				Title:    "Richiesta Patente",
				Components: []discordgo.MessageComponent{
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "cittadino_nome",
								Label:       "Nome e Cognome Cittadino",
								Style:       discordgo.TextInputShort,
								Placeholder: "Mario Rossi",
								Required:    true,
							},
						},
					},
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "veicolo",
								Label:       "Modello Veicolo",
								Style:       discordgo.TextInputShort,
								Placeholder: "Fiat Panda",
								Required:    true,
							},
						},
					},
					discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.TextInput{
								CustomID:    "tipo_patente",
								Label:       "Tipo di Patente",
								Style:       discordgo.TextInputShort,
								Placeholder: "Tipo A / Tipo B",
								Required:    true,
							},
						},
					},
				},
			},
		})
		if err != nil {
			log.Println("Errore nell'invio del modale patente:", err)
		}

	case "turno":
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "🟢 **Turno avviato con successo!** Posizione in servizio: **1**",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})

	case "promuovi":
		target := data.Options[0].UserValue(s)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("⬆️ L'utente **%s** è stato promosso con successo!", target.Username),
			},
		})

	case "retrocedi":
		target := data.Options[0].UserValue(s)
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: fmt.Sprintf("⬇️ L'utente **%s** è stato retrocesso.", target.Username),
			},
		})
	}
}

func handleModalSubmits(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.ModalSubmitData()

	if data.CustomID == "patente_modal" {
		var nome, veicolo, tipo string
		for _, rc := range data.Components {
			row := rc.(*discordgo.ActionsRow)
			input := row.Components[0].(*discordgo.TextInput)
			switch input.CustomID {
			case "cittadino_nome":
				nome = input.Value
			case "veicolo":
				veicolo = input.Value
			case "tipo_patente":
				tipo = input.Value
			}
		}

		embed := &discordgo.MessageEmbed{
			Title:       "📄 Nuova Richiesta Patente",
			Description: fmt.Sprintf("**Cittadino:** %s\n**Veicolo:** %s\n**Tipo:** %s\n**Richiedente:** <@%s>", nome, veicolo, tipo, i.Member.User.ID),
			Color:       0x3498db,
		}

		components := []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Accetta",
						Style:    discordgo.ButtonStyleSuccess,
						CustomID: "patente_accetta",
					},
					discordgo.Button{
						Label:    "Rifiuta",
						Style:    discordgo.ButtonStyleDanger,
						CustomID: "patente_rifiuta",
					},
				},
			},
		}

		_, err := s.ChannelMessageSendComplex(LogChannelID, &discordgo.MessageSend{
			Embeds:     []*discordgo.MessageEmbed{embed},
			Components: components,
		})

		if err != nil {
			log.Println("Errore nell'invio del messaggio di verifica staff:", err)
		}

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "✅ Richiesta della patente inviata con successo allo staff in attesa di verifica!",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	}
}

func handleComponents(s *discordgo.Session, i *discordgo.InteractionCreate) {
	data := i.MessageComponentData()

	if data.CustomID == "patente_accetta" || data.CustomID == "patente_rifiuta" {
		esito := "✅ **ACCETTATA**"
		color := 0x2ecc71
		if data.CustomID == "patente_rifiuta" {
			esito = "❌ **RIFIUTATA**"
			color = 0xe74c3c
		}

		embed := i.Message.Embeds[0]
		embed.Color = color
		embed.Description += fmt.Sprintf("\n\n**Esito:** %s da <@%s>", esito, i.Member.User.ID)

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				Embeds:     []*discordgo.MessageEmbed{embed},
				Components: []discordgo.MessageComponent{},
			},
		})

		channels, err := s.UserChannelCreate(i.Member.User.ID)
		if err == nil {
			s.ChannelMessageSend(channels.ID, fmt.Sprintf("La tua richiesta di patente è stata: %s", esito))
		}
	}
}
