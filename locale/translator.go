package locale

import (
	"embed"
	"fmt"

	"github.com/BurntSushi/toml"
	"github.com/dgf/go-ssr-x/log"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

//go:embed active.*.toml
var localeFS embed.FS

var bundle *i18n.Bundle

type Translator interface {
	Translate(messageID string) string
	TranslateData(messageID string, data map[string]string) string
}

type translator struct {
	localizer *i18n.Localizer
}

var messageByID map[string]*i18n.Message

func init() {
	bundle = i18n.NewBundle(language.English)
	bundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
	if _, err := bundle.LoadMessageFileFS(localeFS, "active.de.toml"); err != nil {
		log.Error("bundle message load failed", err)
	}

	messages := [...]*i18n.Message{
		{ID: "bad_request_path_param", Other: "Bad Request, invalid path param '{{.param}}' value '{{.value}}'"},
		{ID: "client_error", Other: "Client Error"},
		{ID: "conflict_task_update", Other: "The task update has failed due to an editing conflict. Please try the update again."},
		{ID: "database_error", Other: "Database Error {{.message}}"},
		{ID: "internal_server_error", Other: "Internal Server Error"},
		{ID: "not_found_path", Other: "Not Found '{{.method}} {{.path}}'"},
		{ID: "not_found_task", Other: "Task '{{.id}}' not found."},
		{ID: "ok_task_created", Other: "Task '{{.id}}' created."},
		{ID: "ok_task_updated", Other: "Task '{{.id}}' updated."},
		{ID: "page_title", Other: "My Tasks"},
		{ID: "task_add", Other: "add"},
		{ID: "task_back", Other: "back"},
		{ID: "task_cancel", Other: "cancel"},
		{ID: "task_confirm_delete", Other: "Are you sure?"},
		{ID: "task_create", Other: "create"},
		{ID: "task_created_at", Other: "create at"},
		{ID: "task_creating", Other: "creating"},
		{ID: "task_description", Other: "description"},
		{ID: "task_due_date", Other: "due date"},
		{ID: "task_edit", Other: "edit"},
		{ID: "task_order_created_asc", Other: "Oldest (created)"},
		{ID: "task_order_created_desc", Other: "Newest (created)"},
		{ID: "task_order_due_date_asc", Other: "Urgent (due date)"},
		{ID: "task_order_due_date_desc", Other: "Relaxed (due date)"},
		{ID: "task_order_subject_asc", Other: "Subject (alphabetical)"},
		{ID: "task_order_subject_desc", Other: "Subject (reverse)"},
		{ID: "task_save", Other: "save"},
		{ID: "task_saving", Other: "saving"},
		{ID: "task_subject", Other: "subject"},
		{ID: "tasks_loading", Other: "loading"},
	}

	messageByID = make(map[string]*i18n.Message, len(messages))
	for _, m := range messages {
		messageByID[m.ID] = m
	}
}

func RequestTranslator(lang language.Tag) Translator {
	return &translator{
		localizer: i18n.NewLocalizer(bundle, lang.String()),
	}
}

func (m *translator) Translate(messageID string) string {
	if message, ok := messageByID[messageID]; !ok {
		log.Warn("unknown translation message", "messageID", messageID)
		return messageID
	} else if translation, err := m.localizer.Localize(&i18n.LocalizeConfig{DefaultMessage: message}); err != nil {
		log.Warn(fmt.Sprintf("translation error: %v", err), "messageID", messageID)
		return messageID
	} else {
		return translation
	}
}

func (m *translator) TranslateData(messageID string, data map[string]string) string {
	if message, ok := messageByID[messageID]; !ok {
		log.Warn("unknown translation message", "messageID", messageID)
		return messageID
	} else if translation, err := m.localizer.Localize(&i18n.LocalizeConfig{DefaultMessage: message, TemplateData: data}); err != nil {
		log.Warn(fmt.Sprintf("translation error: %v", err), "messageID", messageID)
		return messageID
	} else {
		return translation
	}
}
