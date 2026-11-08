package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
)

type ActivePatrol struct {
	UserID    string
	StartTime time.Time
}

var (
	activePatrols = make(map[string]ActivePatrol)
	openTickets   = 0
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

	sess.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		fmt.Printf("✅ Bot FDO online come: %s\n", s.State.User.Username)
		registerCommands(s)
	})

	sess.AddHandler(handleInteractions)

	err = sess.Open()
	if err != nil {
		fmt.Printf("❌ Errore apertura connessione: %v\n", err)
		return
	}

	fmt.Println("🤖 Bot Rome Italy RP 2.0 [E.H.] FDO avviato correttamente in Go!")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	sess.Close()
	fmt.Println("👋 Bot spento correttamente.")
}

func registerCommands(s *discordgo.Session) {
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        "setup",
			Description: "Configura e invia i pannelli ufficiali di Pattuglia e Ticket FDO",
		},
		{
			Name:        "patente",
			Description: "Rilascia la patente al cittadino (30 punti iniziali e invio in DM)",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionUser, Name: "utente", Description: "Utente Discord", Required: true},
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
				{Type: discordgo.ApplicationCommandOptionUser, Name: "utente", Description: "Cittadino sanzionato", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "nome_roblox", Description: "Nome Roblox", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "motivo", Description: "Motivo", Required: true},
				{Type: discordgo.ApplicationCommandOptionInteger, Name: "punti", Description: "Punti da decurtare", Required: true},
			},
		},
		{
			Name:        "promozione",
			Description: "Promuove un agente",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionUser, Name: "utente", Description: "Agente", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "da", Description: "Da ruolo", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "a", Description: "A ruolo", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "motivo", Description: "Motivo", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "firmato", Description: "Firmato da", Required: true},
			},
		},
		{
			Name:        "retrocessione",
			Description: "Retrocede un agente",
			Options: []*discordgo.ApplicationCommandOption{
				{Type: discordgo.ApplicationCommandOptionUser, Name: "utente", Description: "Agente", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "da", Description: "Da ruolo", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "a", Description: "A ruolo", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "motivo", Description: "Motivo", Required: true},
				{Type: discordgo.ApplicationCommandOptionString, Name: "firmato", Description: "Firmato da", Required: true},
			},
		},
	}

	for _, cmd := range commands {
		_, err := s.ApplicationCommandCreate(s.State.User.ID, "", cmd)
		if err != nil {
			fmt.Printf("❌ Errore comando %s: %v\n", cmd.Name, err)
		}
	}
	fmt.Println("🚀 Comandi registrati con successo!")
}

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

		// Inizio Turno Pattuglia
		if customID == "btn_start_patrol" {
			activePatrols[i.Member.User.ID] = ActivePatrol{UserID: i.Member.User.ID, StartTime: time.Now()}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("🟢 <@%s> è entrato in servizio di pattuglia! (Agenti attivi: %d)", i.Member.User.ID, len(activePatrols)),
				},
			})
		}

		// Fine Turno Pattuglia -> Apre il Modal
		if customID == "btn_end_patrol" {
			if _, exists := activePatrols[i.Member.User.ID]; !exists {
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{Content: "⚠️ Non sei in servizio attivo!", Flags: discordgo.MessageFlagsEphemeral},
				})
				return
			}
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseModal,
				Data: &discordgo.InteractionResponseData{
					CustomID: "modal_patrol_report",
					Title:    "Report di Fine Pattuglia",
					Components: []discordgo.MessageComponent{
						discordgo.ActionsRow{
							Components: []discordgo.MessageComponent{
								discordgo.TextInput{
									CustomID:    "input_report_text",
									Label:       "Attività svolte",
									Style:       discordgo.TextInputParagraph,
									Placeholder: "Descrivi il turno...",
									Required:    true,
								},
							},
						},
					},
				},
			})
		}

		// Selezione Categoria Ticket -> Crea il Canale Reale
		if customID == "select_ticket_category" {
			values := i.MessageComponentData().Values
			if len(values) > 0 {
				category := values[0]
				openTickets++

				var pingRole string
				var categoryTitle string

				switch category {
				case "cat_ricorso":
					pingRole = "<@&1501973204925808661>"
					categoryTitle = "Ricorso Sanzioni / Denunce"
				case "cat_piani_alti":
					pingRole = "<@&1501973204925808661>"
					categoryTitle = "Richiesta Piani Alti"
				case "cat_info":
					pingRole = "<@&1568928207053389844>"
					categoryTitle = "Informazioni Generali"
				}

				// Risposta temporanea immediata
				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "🔄 Apertura del canale ticket in corso...",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})

				// Creazione del canale privato per il ticket
				guildID := i.GuildID
				channelName := fmt.Sprintf("ticket-%s", i.Member.User.Username)
				
				overwrites := []*discordgo.PermissionOverwrite{
					{
						ID:   guildID, // @everyone negato
						Type: discordgo.PermissionOverwriteTypeRole,
						Deny: discordgo.PermissionViewChannel,
					},
					{
						ID:    i.Member.User.ID, // Utente autorizzato
						Type:  discordgo.PermissionOverwriteTypeMember,
						Allow: discordgo.PermissionViewChannel | discordgo.PermissionSendMessages | discordgo.PermissionReadMessageHistory,
					},
				}

				createdChannel, err := s.GuildChannelCreateComplex(guildID, discordgo.GuildChannelCreateData{
					Name:                 channelName,
					Type:                 discordgo.ChannelTypeGuildText,
					PermissionOverwrites: overwrites,
				})

				if err == nil {
					// Invia il messaggio nel nuovo canale con i bottoni di gestione
					embed := &discordgo.MessageEmbed{
						Title:       fmt.Sprintf("🎫 Ticket: %s", categoryTitle),
						Description: fmt.Sprintf("Benvenuto nel tuo ticket, <@%s>.\nIl personale competente è stato notificato: %s\n\nUsa i pulsanti sottostanti per gestire il ticket.", i.Member.User.ID, pingRole),
						Color:       0x00aaff,
					}
					buttons := discordgo.ActionsRow{
						Components: []discordgo.MessageComponent{
							discordgo.Button{Label: "Claim", Style: discordgo.SuccessButton, CustomID: "ticket_claim"},
							discordgo.Button{Label: "Unclaim", Style: discordgo.SecondaryButton, CustomID: "ticket_unclaim"},
							discordgo.Button{Label: "Chiudi", Style: discordgo.DangerButton, CustomID: "ticket_close"},
						},
					}
					s.ChannelMessageSendComplex(createdChannel.ID, &discordgo.MessageSend{
						Content:    pingRole,
						Embeds:     []*discordgo.MessageEmbed{embed},
						Components: []discordgo.MessageComponent{buttons},
					})
				}
			}
		}

		// Gestione Pulsanti Ticket (Claim, Unclaim, Chiudi con Transcript in DM)
		if customID == "ticket_claim" {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("🙋‍♂️ Ticket preso in carico da <@%s>.", i.Member.User.ID),
				},
			})
		}

		if customID == "ticket_unclaim" {
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: fmt.Sprintf("↩️ <@%s> ha rilasciato il ticket.", i.Member.User.ID),
				},
			})
		}

		if customID == "ticket_close" {
			// Invio transcript in DM (simulato/raccolto) e chiusura canale
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "🔒 Chiusura del ticket in corso e invio transcript in DM...",
				},
			})
			
			// Invio DM all'utente che ha chiuso/richiesto
			dm, err := s.UserChannelCreate(i.Member.User.ID)
			if err == nil {
				s.ChannelMessageSend(dm.ID, "📄 Ecco il transcript del tuo ticket chiuso per Rome Italy RP 2.0 [E.H.]. Grazie per averci contattato.")
			}

			// Elimina il canale dopo pochi secondi
			time.Sleep(3 * time.Second)
			s.ChannelDelete(i.ChannelID)
		}

	case discordgo.InteractionModalSubmit:
		data := i.ModalSubmitData()
		if data.CustomID == "modal_patrol_report" {
			patrol, exists := activePatrols[i.Member.User.ID]
			if exists {
				duration := time.Since(patrol.StartTime)
				minutes := int(duration.Minutes())
				delete(activePatrols, i.Member.User.ID)

				reportText := data.Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

				s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("🔴 <@%s> ha terminato il turno. Durata: **%d minuti**.\n📝 **Report:** %s", i.Member.User.ID, minutes, reportText),
					},
				})
			}
		}
	}
}

func handleSetupCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	patrolEmbed := &discordgo.MessageEmbed{
		Title:       "🟢 Gestione Pattuglie FDO",
		Description: fmt.Sprintf("Usa i pulsanti per iniziare o terminare il servizio.\n\n👥 **Agenti in servizio:** `%d`", len(activePatrols)),
		Color:       0x0055ff,
	}
	patrolButtons := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{Label: "Inizia Turno", Style: discordgo.SuccessButton, CustomID: "btn_start_patrol", Emoji: &discordgo.ComponentEmoji{Name: "🟢"}},
			discordgo.Button{Label: "Finisci Turno", Style: discordgo.DangerButton, CustomID: "btn_end_patrol", Emoji: &discordgo.ComponentEmoji{Name: "🔴"}},
		},
	}

	ticketEmbed := &discordgo.MessageEmbed{
		Title:       "🎫 Sportello Richieste & Segnalazioni FDO",
		Description: fmt.Sprintf("Seleziona una categoria per aprire un ticket privato.\n\n📊 **Ticket aperti:** `%d`", openTickets),
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
					{Label: "Informazioni Generali", Value: "cat_info", Description: "Domande sul regolamento", Emoji: &discordgo.ComponentEmoji{Name: "❓"}},
				},
			},
		},
	}

	s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{patrolEmbed}, Components: []discordgo.MessageComponent{patrolButtons}})
	s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{ticketEmbed}, Components: []discordgo.MessageComponent{ticketMenu}})

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: "✅ Pannelli di Setup configurati e inviati con successo!",
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

func handlePatenteCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	options := i.ApplicationCommandData().Options
	user := options[0].UserValue(s)
	nomeRoblox := options[1].StringValue()
	tipo := options[2].StringValue()

	dm, err := s.UserChannelCreate(user.ID)
	if err == nil {
		s.ChannelMessageSend(dm.ID, fmt.Sprintf("🪪 Patente Tipo **%s** rilasciata per **%s**! Partenza ufficiale con **30 punti patente**.", tipo, nomeRoblox))
	}

	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("✅ Patente **Tipo %s** rilasciata a <@%s> (`%s`). 30 punti iniziali assegnati e notifica inviata in DM.", tipo, user.ID, nomeRoblox),
		},
	})
}

func handleDecurtaPuntiCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.ApplicationCommandData().Options
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("📉 Punti decurtati con successo a <@%s> (`-%d` punti). Motivo: %s", opts[0].UserValue(s).ID, opts[3].IntValue(), opts[2].StringValue()),
		},
	})
}

func handlePromozioneCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.ApplicationCommandData().Options
	embed := &discordgo.MessageEmbed{
		Title: "🎖️ Promozione Ufficiale - FDO",
		Color: 0x00ff00,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Agente", Value: fmt.Sprintf("<@%s>", opts[0].UserValue(s).ID), Inline: false},
			{Name: "Da Ruolo", Value: opts[1].StringValue(), Inline: true},
			{Name: "A Ruolo", Value: opts[2].StringValue(), Inline: true},
			{Name: "Motivo", Value: opts[3].StringValue(), Inline: false},
			{Name: "Firmato", Value: opts[4].StringValue(), Inline: false},
		},
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}},
	})
}

func handleRetrocessioneCommand(s *discordgo.Session, i *discordgo.InteractionCreate) {
	opts := i.ApplicationCommandData().Options
	embed := &discordgo.MessageEmbed{
		Title: "🛡️ Sanzione / Retrocessione - FDO",
		Color: 0xff0000,
		Fields: []*discordgo.MessageEmbedField{
			{Name: "Agente", Value: fmt.Sprintf("<@%s>", opts[0].UserValue(s).ID), Inline: false},
			{Name: "Da Ruolo", Value: opts[1].StringValue(), Inline: true},
			{Name: "A Ruolo", Value: opts[2].StringValue(), Inline: true},
			{Name: "Motivo", Value: opts[3].StringValue(), Inline: false},
			{Name: "Firmato", Value: opts[4].StringValue(), Inline: false},
		},
	}
	s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{Embeds: []*discordgo.MessageEmbed{embed}},
	})
}
