package api

import "orbit/internal/discord"

type methodInvoker func(handler *Handler, arguments methodArguments) (interface{}, error)

var methodRegistry = map[string]methodInvoker{
	"get_boards": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.GetBoards()
	},
	"create_board": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.CreateBoard(arguments.stringAt(0), arguments.stringAt(1))
	},
	"update_board": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.UpdateBoard(arguments.int64At(0), arguments.fieldsAt(1))
	},
	"delete_board": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.DeleteBoard(arguments.int64At(0))
	},
	"get_board": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.GetBoard(arguments.int64At(0))
	},
	"restore_board": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.RestoreBoard(arguments.int64At(0), arguments.fieldsAt(1))
	},
	"create_list": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.CreateList(arguments.int64At(0), arguments.stringAt(1))
	},
	"rename_list": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.RenameList(arguments.int64At(0), arguments.stringAt(1))
	},
	"delete_list": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.DeleteList(arguments.int64At(0))
	},
	"move_list": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.MoveList(arguments.int64At(0), arguments.intAt(1))
	},
	"create_card": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.CreateCard(arguments.int64At(0), arguments.stringAt(1))
	},
	"update_card": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.UpdateCard(arguments.int64At(0), arguments.fieldsAt(1))
	},
	"delete_card": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.DeleteCard(arguments.int64At(0))
	},
	"move_card": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.MoveCard(arguments.int64At(0), arguments.int64At(1), arguments.intAt(2))
	},
	"move_cards": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.MoveCards(arguments.int64SliceAt(0), arguments.int64At(1), arguments.intAt(2))
	},
	"move_lists": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.MoveLists(arguments.int64SliceAt(0), arguments.intAt(1))
	},
	"batch_update_cards": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.BatchUpdateCards(arguments.int64SliceAt(0), arguments.fieldsAt(1))
	},
	"batch_delete_cards": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.BatchDeleteCards(arguments.int64SliceAt(0))
	},
	"duplicate_cards": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.DuplicateCards(arguments.int64SliceAt(0))
	},
	"batch_delete_lists": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.BatchDeleteLists(arguments.int64SliceAt(0))
	},
	"duplicate_lists": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.DuplicateLists(arguments.int64SliceAt(0))
	},
	"add_checklist": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.AddChecklist(arguments.int64At(0), arguments.stringAt(1))
	},
	"rename_checklist": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.RenameChecklist(arguments.int64At(0), arguments.stringAt(1))
	},
	"move_checklist": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.MoveChecklist(arguments.int64At(0), arguments.intAt(1))
	},
	"delete_checklist": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.DeleteChecklist(arguments.int64At(0))
	},
	"add_checklist_item": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.AddChecklistItem(arguments.int64At(0), arguments.stringAt(1))
	},
	"toggle_checklist_item": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.ToggleChecklistItem(arguments.int64At(0), arguments.boolAt(1))
	},
	"edit_checklist_item": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.EditChecklistItem(arguments.int64At(0), arguments.stringAt(1))
	},
	"delete_checklist_item": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.DeleteChecklistItem(arguments.int64At(0))
	},
	"move_checklist_item": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.MoveChecklistItem(arguments.int64At(0), arguments.int64At(1), arguments.intAt(2))
	},
	"add_attachments_dialog": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.AddAttachmentsDialog(arguments.int64At(0))
	},
	"open_attachment": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.OpenAttachment(arguments.int64At(0))
	},
	"save_attachment_as": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.SaveAttachmentAs(arguments.int64At(0))
	},
	"rename_attachment": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.RenameAttachment(arguments.int64At(0), arguments.stringAt(1))
	},
	"delete_attachment": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.DeleteAttachment(arguments.int64At(0))
	},
	"set_active_board": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.SetActiveBoard(arguments.int64At(0))
	},
	"set_title": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.SetTitle(arguments.stringAt(0))
	},
	"get_fonts": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.GetFonts()
	},
	"get_discord_status": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.GetDiscordStatus()
	},
	"get_settings": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.GetSettings()
	},
	"set_theme": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.SetTheme(arguments.stringAt(0))
	},
	"set_discord_enabled": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.SetDiscordEnabled(arguments.boolAt(0))
	},
	"set_tooltips_enabled": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.SetTooltipsEnabled(arguments.boolAt(0))
	},
	"toggle_discord_rpc": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.ToggleDiscordRPC()
	},
	"reset_application_data": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.ResetApplicationData()
	},
	"import_boards_dialog": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		return handler.ImportBoardsDialog()
	},
	"set_presence_context": func(handler *Handler, arguments methodArguments) (interface{}, error) {
		var presenceContext discord.Context
		if !arguments.decode(0, &presenceContext) {
			presenceContext = discord.Context{}
		}
		return handler.SetPresenceContext(presenceContext)
	},
}
