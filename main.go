package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

// Struttura per tracciare lo stato dei turni e dei ticket in memoria
type ActivePatrol struct {
	UserID    string
	StartTime time.Time
	Vehicle   string
}

var (
	activePatrols = make(map[string]ActivePatrol) // Mappa per calcolare le ore di turno
	openTickets   = 0                            // Contatore globale dei ticket aperti
)

func main() {
	token := os.Getenv("TOKEN")
	if token == "" {
		fmt.Println("❌ Errore: Token Discord non trovato nelle variabili d'ambiente!")
		return
	}

	// 🌐 Server HTTP Keep-Alive per Render
	port := os.Getenv("PORT")
	if port == "" {
		port = "3000"
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Bot Rome Italy FDO Online 24/7! 🚀"))
	})
	go func() {
		fmt.Println("🌐 Server HTTP avviato sulla porta " + port)
		http.ListenAndServe(":"+port, nil)
	}()

	// 🤖 Inizializzazione Bot Discord
	sess, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Printf("❌ Errore creazione sessione: %v\n", err)
		return
	}

	sess.Identify.Intents = discordgo.IntentsGuilds |
		discordgo.IntentsGuildMessages |
		discordgo.IntentsMessageContent |
		discordgo.IntentsGuildMembers

	// Evento Ready & Registrazione Comandi Slash
	sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Printf("✅ Bot FDO online come: %s\n", s.State.User.Username)
		registerCommands(s)
	})

	// Gestore delle Interazioni (Slash Commands, Bottoni, Select Menus, Modals)
	sess.AddHandler(handleInteractions)

	// Apertura connessione WebSocket
	err = sess.Open()
	if err != nil {
		fmt.Printf("❌ Errore apertura connessione: %v\n", err)
		return
	}

	fmt.Println("🤖 Bot Rome Italy RP 2.0 [E.H.] FDO avviato correttamente in Go! Premi CTRL-C per uscire.")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	sess.Close()
	fmt.Println("👋 Bot spento correttamente.")
}

// 📌 Definizione e Registrazione dei Comandi Slash
func registerCommands(s *discordgo.Session) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "setup",
			Description: "Invia i pannelli ufficiali (Pattuglia & Ticket FDO) nel canale corrente",
		},
		{
			Name:        "patente",
			Description: "Rilascia la patente al cittadino (Assegna 30 punti iniziali e invia in DM)",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionUser, Name: "utente", Description: "L'utente Discord", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "nome_roblox", Description: "Nome Roblox", Required: true},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "tipo",
					Description: "Tipo di patente",
					Required:    true,
					Choices: []*discordgo.ApplicationCommandOptionChoice{
						{Name: "Patente A", Value: "A"},
						{Name: "Patente B", Value: "B"},
						{Name: "Patente C", Value: "C"},
						{Name: "Patente D", Value: "D"},
					},
				},
			},
		},
		{
			Name:        "decurta-punti",
			Description: "Decurta i punti dalla patente di un cittadino",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionUser, Name: "utente", Description: "Il cittadino sanzionato", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "nome_roblox", Description: "Nome Roblox", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "motivo", Description: "Motivo della decurtazione", Required: true},
				{Type: discordgo.ApplicationCommandOptionInteger, Name: "punti", Description: "Punti da decurtare", Required: true},
			},
		},
		{
			Name:        "promozione",
			Description: "Promuove un agente registrando l'atto ufficiale",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionUser, Name: "utente", Description: "Agente promosso", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "da", Description: "Ruolo precedente", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "a", Description: "Nuovo ruolo", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "motivo", Description: "Motivo", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "firmato", Description: "Chi firma l'atto", Required: true},
			},
		},
		{
			Name:        "retrocessione",
			Description: "Retrocede un agente o applica sanzione disciplinare",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionUser, Name: "utente", Description: "Agente interessato", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "da", Description: "Ruolo precedente", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "a", Description: "Nuovo ruolo", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "motivo", Description: "Motivo", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "firmato", Description: "Chi firma l'atto", Required: true},
			},
		},
	}

	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			fmt.Printf("❌ Errore nella creazione del comando %s: %v\n", cmd.Name, err)
		}
	}
	fmt.Println("🚀 Tutti i comandi FDO sono stati registrati con successo!")
}

// 🎛️ Gestore Centrale delle Interazioni
func handleInteractions(s *discordgo.Session, i *discordgo.InteractionCreate) {
	switch i.Type {
	case discordgo.InteractionApplicationCommand:
		data := i.ApplicationCommandData()
		switch data.Name {
		case "setup":
			handleSetupCommand(s, i)
		case "patente":
			handlePatenteCommand(s, i)
		case "decurta-punti":
			handleDecurtaPuntiCommand(s, i)
		case "promozione":
			handlePromozioneCommand(s, i)
		case "retrocessione":
			handleRetrocessioneCommand(s, i)
		}

	case discordgo.InteractionMessageComponent:
		customID := i.MessageComponentData().CustomID

		// Gestione Pulsante Inizio Pattuglia
		if customID == "btn_start_patrol" {
			activePatrols[i.Member.User.ID] = ActivePatrol{
				UserID:    i.Member.User.ID,
				StartTime: time.Now(),
				Vehicle:   "Alfa Romeo / Pattuglia",
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("🟢 Il signor <@%s> è in servizio di pattuglia! (Agenti operativi: %d)", i.Member.User.ID, len(activePatrols)),
				},
			})
		}

		// Gestione Pulsante Fine Pattuglia (Apre il Modal per il report)
		if customID == "btn_end_patrol" {
			if _, exists := activePatrols[i.Member.User.ID]; !exists {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "⚠️ Non risulti in servizio di pattuglia attivo!",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				return
			}

			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID: "modal_patrol_report",
					Title:    "Report di Fine Pattuglia",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "input_report_text",
									Label:       "Descrivi le attività svolte durante il turno",
									Style:       discordgo.TextInputParagraph,
									Placeholder: "Es. Posti di blocco, interventi, scorte...",
									Required:    true,
									MaxLength:   1000,
								},
							},
						},
					},
				},
			})
			if err != nil {
				fmt.Printf("Errore apertura modal: %v\n", err)
			}
		}

		// Gestione Apertura Ticket tramite Menu a Tendina
		if customID == "select_ticket_category" {
			values := i.MessageComponentData().Values
			if len(values) > 0 {
				category := values[0]
				openTickets++

				var pingRole string
				var categoryName string

				switch category {
				case "cat_ricorso":
					pingRole = "<@&1501973204925808661>"
					categoryName = "Ricorso Sanzioni / Denunce"
				case "cat_piani_alti":
					pingRole = "<@&1501973204925808661>"
					categoryName = "Richiesta Piani Alti"
				case "cat_info":
					pingRole = "<@&1568928207053389844>"
					categoryName = "Informazioni Generali"
				}

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("✅ Ticket aperto con successo per **%s**! (Ruolo notificato: %s)", categoryName, pingRole),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
			}
		}

	case discordgo.InteractionModalSubmit:
		data := i.ModalSubmitData()
		if data.CustomID == "modal_patrol_report" {
			patrol, exists := activePatrols[i.Member.User.ID]
			if exists {
				duration := time.Since(patrol.StartTime)
				minutes := int(duration.Minutes())
				hours := minutes / 60
				remMins := minutes % 60

				reportText := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value
				delete(activePatrols, i.Member.User.ID)

				durationStr := fmt.Sprintf("%d minuti", minutes)
				if hours > 0 {
					durationStr = fmt.Sprintf("%d ore e %d minuti", hours, remMins)
				}

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("🔴 Il signor <@%s> ha terminato il turno con ben **%s** di attività.\n📝 **Report:** %s", i.Member.User.ID, durationStr, reportText),
					},
				})
			}
		}
	}
}

// ⚙️ Comando Setup: Invia i pannelli di Pattuglia e Ticket nel canale
func handleSetupCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Pannello Pattuglia con contatore
	patrolEmbed := &discordgo.MessageEmbed{
		Title:       "🟢 Gestione Pattuglie FDO",
		Description: fmt.Sprintf("Usa i pulsanti sottostanti per iniziare o terminare il servizio.\n\n👥 **Agenti attualmente in servizio:** `%d`", len(activePatrols)),
		Color:       0x0055ff,
	}
	patrolButtons := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{Label: "Inizia Turno", Style: discordgo.SuccessButton, CustomID: "btn_start_patrol", Emoji: &discordgo.ComponentEmoji{Name: "🟢"}},
			discordgo.Button{Label: "Finisci Turno", Style: discordgo.DangerButton, CustomID: "btn_end_patrol", Emoji: &discordgo.ComponentEmoji{Name: "🔴"}},
		},
	}

	// Pannello Ticket FDO con le 3 categorie e contatore ticket aperti
	ticketEmbed := &discordgo.MessageEmbed{
		Title:       "🎫 Sportello Richieste & Segnalazioni FDO",
		Description: fmt.Sprintf("Hai bisogno di metterti in contatto con il comando? Seleziona una categoria dal menu.\n\n📊 **Ticket aperti:** `%d`", openTickets),
		Color:       0xffaa00,
	}
	ticketMenu := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.SelectMenu{
				CustomID:    "select_ticket_category",
				Placeholder: "Seleziona il motivo del ticket...",
				Options: []discordgo.SelectMenuOption{
					{Label: "Ricorso Sanzioni / Denunce", Value: "cat_ricorso", Description: "Contesta sanzioni o sporgi denuncia", Emoji: &discordgo.ComponentEmoji{Name: "⚖️"}},
					{Label: "Richiesta Piani Alti", Value: "cat_piani_alti", Description: "Contatta l'Alto Comando", Emoji: &discordgo.ComponentEmoji{Name: "👑"}},
					{Label: "Informazioni Generali", Value: "cat_info", Description: "Domande sul regolamento o generale", Emoji: &discordgo.ComponentEmoji{Name: "❓"}},
				},
			},
		},
	}

	// Invio dei pannelli nel canale
	s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{patrolEmbed},
		Components: []discordgo.MessageComponent{patrolButtons},
	})

	s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
		Embeds:     []*discordgo.MessageEmbed{ticketEmbed},
		Components: []discordgo.MessageComponent{ticketMenu},
	})

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "✅ Pannelli di Setup FDO inviati con successo in questo canale!",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// 🪪 Comando Patente: Rilascio automatico, 30 punti e DM
func handlePatenteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	user := options[0].UserValue(s)
	nomeRoblox := options[1].StringValue()
	tipoPatente := options[2].StringValue()

	// Invio notifica in DM all'utente
	dmChannel, err := s.UserChannelCreate(user.ID)
	if err == nil {
		s.ChannelMessageSend(dmChannel.ID, fmt.Sprintf("🪪 **Congratulazioni!** La tua Patente **Tipo %s** per il cittadino **%s** è stata rilasciata con successo.\n🏆 Parti ufficialmente con **30 punti patente** iniziali!", tipoPatente, nomeRoblox))
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("✅ Patente **Tipo %s** rilasciata a <@%s> (Roblox: `%s`). Assegnati 30 punti iniziali e inviato messaggio in DM!", tipoPatente, user.ID, nomeRoblox),
		},
	})
}

// 📉 Comando Decurta Punti
func handleDecurtaPuntiCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	user := options[0].UserValue(s)
	nomeRoblox := options[1].StringValue()
	motivo := options[2].StringValue()
	punti := options[3].IntValue()

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("📉 **Decurtazione Punti Patente**\n👤 Cittadino: <@%s> (`%s`)\n❌ Punti scalati: `-%d`\n📋 Motivo: %s", user.ID, nomeRoblox, punti, motivo),
		},
	})
}

// 🎖️ Comando Promozione
func handlePromozioneCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	user := options[0].UserValue(s)
	daRuolo := options[1].StringValue()
	aRuolo := options[2].StringValue()
	motivo := options[3].StringValue()
	firmato := options[4].StringValue()

	embed := &discordgo.MessageEmbed{
		Title:       "🎖️ Atto Ufficiale di Promozione - FDO",
		Description: fmt.Sprintf("L'operatore <@%s> è stato ufficialmente promosso.", user.ID),
		Color:       0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Da Ruolo", Value: daRuolo, Inline: true},
			{Name: "A Ruolo", Value: aRuolo, Inline: true},
			{Name: "Motivo", Value: motivo, Inline: false},
			{Name: "Firmato", Value: firmato, Inline: false},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// 🛡️ Comando Retrocessione
func handleRetrocessioneCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	user := options[0].UserValue(s)
	daRuolo := options[1].StringValue()
	aRuolo := options[2].StringValue()
	motivo := options[3].StringValue()
	firmato := options[4].StringValue()

	embed := &discordgo.MessageEmbed{
		Title:       "🛡️ Sanzione / Retrocessione - FDO",
		Description: fmt.Sprintf("Provvedimento disciplinare emesso nei confronti di <@%s>.", user.ID),
		Color:       0xff0000,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Da Ruolo", Value: daRuolo, Inline: true},
			{Name: "A Ruolo", Value: aRuolo, Inline: true},
			{Name: "Motivo", Value: motivo, Inline: false},
			{Name: "Firmato", Value: firmato, Inline: false},
		},
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}
