package event

// Определение общих типов событий в движке

const (
	// События жизненного цикла приложения
	EventAppInit     EventType = "app.init"
	EventAppStart    EventType = "app.start"
	EventAppStop     EventType = "app.stop"
	EventAppPause    EventType = "app.pause"
	EventAppResume   EventType = "app.resume"
	EventAppShutdown EventType = "app.shutdown"

	// События окна
	EventWindowCreate  EventType = "window.create"
	EventWindowClose   EventType = "window.close"
	EventWindowResize  EventType = "window.resize"
	EventWindowFocus   EventType = "window.focus"
	EventWindowUnfocus EventType = "window.unfocus"

	// События ввода - клавиатура
	EventKeyPress   EventType = "input.key.press"
	EventKeyRelease EventType = "input.key.release"
	EventKeyRepeat  EventType = "input.key.repeat"

	// События ввода - мышь
	EventMouseMove        EventType = "input.mouse.move"
	EventMouseButtonPress EventType = "input.mouse.button.press"
	EventMouseButtonRelease EventType = "input.mouse.button.release"
	EventMouseScroll      EventType = "input.mouse.scroll"
	EventMouseEnter       EventType = "input.mouse.enter"
	EventMouseLeave       EventType = "input.mouse.leave"

	// События сущностей
	EventEntityCreate  EventType = "entity.create"
	EventEntityDestroy EventType = "entity.destroy"

	// События компонентов
	EventComponentAdd    EventType = "component.add"
	EventComponentRemove EventType = "component.remove"

	// События рендеринга
	EventRenderBegin EventType = "render.begin"
	EventRenderEnd   EventType = "render.end"
	EventFrameBegin  EventType = "frame.begin"
	EventFrameEnd    EventType = "frame.end"

	// События загрузки ресурсов
	EventResourceLoad   EventType = "resource.load"
	EventResourceUnload EventType = "resource.unload"
	EventResourceError  EventType = "resource.error"

	// События коллизий
	EventCollisionEnter EventType = "collision.enter"
	EventCollisionExit  EventType = "collision.exit"
	EventCollisionStay  EventType = "collision.stay"

	// События UI
	EventUIClick   EventType = "ui.click"
	EventUIHover   EventType = "ui.hover"
	EventUIChange  EventType = "ui.change"
	EventUISubmit  EventType = "ui.submit"

	// События игровой логики (RPG)
	EventPlayerDamage    EventType = "game.player.damage"
	EventPlayerHeal      EventType = "game.player.heal"
	EventPlayerLevelUp   EventType = "game.player.levelup"
	EventEnemySpawn      EventType = "game.enemy.spawn"
	EventEnemyDeath      EventType = "game.enemy.death"
	EventItemPickup      EventType = "game.item.pickup"
	EventItemDrop        EventType = "game.item.drop"
	EventQuestStart      EventType = "game.quest.start"
	EventQuestComplete   EventType = "game.quest.complete"
	EventDialogueStart   EventType = "game.dialogue.start"
	EventDialogueEnd     EventType = "game.dialogue.end"
)

// WindowResizeData данные события изменения размера окна
type WindowResizeData struct {
	Width  int
	Height int
}

// KeyEventData данные события клавиатуры
type KeyEventData struct {
	Key      int
	Scancode int
	Action   int
	Mods     int
}

// MouseMoveData данные события движения мыши
type MouseMoveData struct {
	X     float64
	Y     float64
	DeltaX float64
	DeltaY float64
}

// MouseButtonData данные события кнопки мыши
type MouseButtonData struct {
	Button int
	Action int
	Mods   int
	X      float64
	Y      float64
}

// MouseScrollData данные события прокрутки мыши
type MouseScrollData struct {
	XOffset float64
	YOffset float64
}

// CollisionData данные события коллизии
type CollisionData struct {
	EntityA uint64
	EntityB uint64
}

// ResourceLoadData данные события загрузки ресурса
type ResourceLoadData struct {
	Path string
	Type string
}

// ResourceErrorData данные события ошибки ресурса
type ResourceErrorData struct {
	Path  string
	Error error
}

// DamageData данные события получения урона
type DamageData struct {
	EntityID   uint64
	Amount     float32
	DamageType string
	Source     uint64
}

// HealData данные события исцеления
type HealData struct {
	EntityID uint64
	Amount   float32
	Source   uint64
}
