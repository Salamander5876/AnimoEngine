package rpg

import (
	"github.com/Salamander5876/AnimoEngine/pkg/core/ecs"
)

// RegenerationSystem система регенерации здоровья, маны и выносливости
type RegenerationSystem struct {
	ecs.BaseSystem
}

// NewRegenerationSystem создает новую систему регенерации
func NewRegenerationSystem() *RegenerationSystem {
	return &RegenerationSystem{
		BaseSystem: ecs.NewBaseSystem(10), // Низкий приоритет
	}
}

// Update обновляет регенерацию
func (s *RegenerationSystem) Update(deltaTime float32, em *ecs.EntityManager) {
	entities := em.GetAllEntities()

	for _, entityID := range entities {
		// Регенерация здоровья
		if em.HasComponent(entityID, HealthComponentType) {
			comp, _ := em.GetComponent(entityID, HealthComponentType)
			health := comp.(*HealthComponent)

			if health.Current < health.Max && health.Regeneration > 0 {
				health.Heal(health.Regeneration * deltaTime)
			}
		}

		// Регенерация маны
		if em.HasComponent(entityID, ManaComponentType) {
			comp, _ := em.GetComponent(entityID, ManaComponentType)
			mana := comp.(*ManaComponent)

			if mana.Current < mana.Max && mana.Regeneration > 0 {
				mana.RestoreMana(mana.Regeneration * deltaTime)
			}
		}

		// Регенерация выносливости
		if em.HasComponent(entityID, StaminaComponentType) {
			comp, _ := em.GetComponent(entityID, StaminaComponentType)
			stamina := comp.(*StaminaComponent)

			if stamina.Current < stamina.Max && stamina.Regeneration > 0 {
				stamina.RestoreStamina(stamina.Regeneration * deltaTime)
			}
		}
	}
}

// CombatSystem простая система боя
type CombatSystem struct {
	ecs.BaseSystem
	attackQueue []AttackAction
}

// AttackAction действие атаки
type AttackAction struct {
	AttackerID ecs.EntityID
	TargetID   ecs.EntityID
	Damage     float32
	DamageType string
}

// NewCombatSystem создает новую боевую систему
func NewCombatSystem() *CombatSystem {
	return &CombatSystem{
		BaseSystem:  ecs.NewBaseSystem(5), // Средний приоритет
		attackQueue: make([]AttackAction, 0),
	}
}

// QueueAttack добавляет атаку в очередь
func (s *CombatSystem) QueueAttack(action AttackAction) {
	s.attackQueue = append(s.attackQueue, action)
}

// Update обрабатывает атаки
func (s *CombatSystem) Update(deltaTime float32, em *ecs.EntityManager) {
	// Обрабатываем все атаки в очереди
	for _, action := range s.attackQueue {
		s.processAttack(action, em)
	}

	// Очищаем очередь
	s.attackQueue = s.attackQueue[:0]
}

// processAttack обрабатывает одну атаку
func (s *CombatSystem) processAttack(action AttackAction, em *ecs.EntityManager) {
	// Проверяем наличие атакующего
	if !em.Exists(action.AttackerID) {
		return
	}

	// Проверяем наличие цели
	if !em.Exists(action.TargetID) {
		return
	}

	// Получаем компонент здоровья цели
	if !em.HasComponent(action.TargetID, HealthComponentType) {
		return
	}

	healthComp, _ := em.GetComponent(action.TargetID, HealthComponentType)
	health := healthComp.(*HealthComponent)

	// Вычисляем урон с учетом характеристик атакующего
	finalDamage := action.Damage

	if em.HasComponent(action.AttackerID, StatsComponentType) {
		statsComp, _ := em.GetComponent(action.AttackerID, StatsComponentType)
		stats := statsComp.(*StatsComponent)

		// Добавляем урон от характеристик
		if action.DamageType == "physical" {
			finalDamage += stats.GetPhysicalDamage()
		} else if action.DamageType == "magical" {
			finalDamage += stats.GetMagicalDamage()
		}

		// Проверка критического удара
		if s.rollCritical(stats.GetCriticalChance()) {
			finalDamage *= 2.0
		}
	}

	// Наносим урон
	health.Damage(finalDamage)

	// Проверяем смерть
	if health.IsDead() {
		s.onEntityDeath(action.TargetID, action.AttackerID, em)
	}
}

// rollCritical проверяет выпадение критического удара
func (s *CombatSystem) rollCritical(chance float32) bool {
	// Простая генерация случайного числа (в продакшене использовать crypto/rand)
	// return rand.Float32() < chance
	return false // Заглушка
}

// onEntityDeath обрабатывает смерть сущности
func (s *CombatSystem) onEntityDeath(deadID, killerID ecs.EntityID, em *ecs.EntityManager) {
	// Начисляем опыт убийце
	if em.HasComponent(killerID, StatsComponentType) {
		statsComp, _ := em.GetComponent(killerID, StatsComponentType)
		stats := statsComp.(*StatsComponent)

		// Простая формула опыта
		expGain := 50 // Базовое значение
		if stats.AddExperience(expGain) {
			// Произошло повышение уровня
			// Здесь можно отправить событие
		}
	}

	// Здесь можно добавить дроп предметов, анимацию смерти и т.д.
}

// LevelScalingSystem система масштабирования характеристик от уровня
type LevelScalingSystem struct {
	ecs.BaseSystem
}

// NewLevelScalingSystem создает систему масштабирования
func NewLevelScalingSystem() *LevelScalingSystem {
	return &LevelScalingSystem{
		BaseSystem: ecs.NewBaseSystem(15),
	}
}

// Update пересчитывает характеристики на основе уровня
func (s *LevelScalingSystem) Update(deltaTime float32, em *ecs.EntityManager) {
	entities := em.GetAllEntities()

	for _, entityID := range entities {
		// Пропускаем сущности без характеристик
		if !em.HasComponent(entityID, StatsComponentType) {
			continue
		}

		statsComp, _ := em.GetComponent(entityID, StatsComponentType)
		stats := statsComp.(*StatsComponent)

		// Обновляем максимальное здоровье на основе живучести
		if em.HasComponent(entityID, HealthComponentType) {
			healthComp, _ := em.GetComponent(entityID, HealthComponentType)
			health := healthComp.(*HealthComponent)

			newMaxHP := float32(100 + stats.Vitality*10)
			if health.Max != newMaxHP {
				// Сохраняем процент здоровья
				percent := health.GetHealthPercent()
				health.Max = newMaxHP
				health.Current = health.Max * percent
			}
		}

		// Обновляем максимальную ману на основе интеллекта
		if em.HasComponent(entityID, ManaComponentType) {
			manaComp, _ := em.GetComponent(entityID, ManaComponentType)
			mana := manaComp.(*ManaComponent)

			newMaxMana := float32(50 + stats.Intelligence*5)
			if mana.Max != newMaxMana {
				percent := mana.Current / mana.Max
				mana.Max = newMaxMana
				mana.Current = mana.Max * percent
			}
		}

		// Обновляем максимальную выносливость на основе ловкости
		if em.HasComponent(entityID, StaminaComponentType) {
			staminaComp, _ := em.GetComponent(entityID, StaminaComponentType)
			stamina := staminaComp.(*StaminaComponent)

			newMaxStamina := float32(100 + stats.Agility*8)
			if stamina.Max != newMaxStamina {
				percent := stamina.Current / stamina.Max
				stamina.Max = newMaxStamina
				stamina.Current = stamina.Max * percent
			}
		}
	}
}

// InventorySystem система управления инвентарем
type InventorySystem struct {
	ecs.BaseSystem
}

// NewInventorySystem создает систему инвентаря
func NewInventorySystem() *InventorySystem {
	return &InventorySystem{
		BaseSystem: ecs.NewBaseSystem(20),
	}
}

// Update обновляет инвентарь (обычно не требует постоянного обновления)
func (s *InventorySystem) Update(deltaTime float32, em *ecs.EntityManager) {
	// Инвентарь обычно обновляется через события или прямые вызовы
	// Здесь можно добавить проверку веса, автоматическую сортировку и т.д.
}

// TransferItem переносит предмет между инвентарями
func (s *InventorySystem) TransferItem(fromID, toID ecs.EntityID, itemID string, quantity int, em *ecs.EntityManager) bool {
	if !em.HasComponent(fromID, InventoryComponentType) ||
		!em.HasComponent(toID, InventoryComponentType) {
		return false
	}

	fromInvComp, _ := em.GetComponent(fromID, InventoryComponentType)
	fromInv := fromInvComp.(*InventoryComponent)

	toInvComp, _ := em.GetComponent(toID, InventoryComponentType)
	toInv := toInvComp.(*InventoryComponent)

	// Проверяем наличие предмета у отправителя
	if !fromInv.HasItem(itemID, quantity) {
		return false
	}

	// Пытаемся добавить получателю
	if !toInv.AddItem(itemID, quantity) {
		return false // Нет места
	}

	// Удаляем у отправителя
	fromInv.RemoveItem(itemID, quantity)
	return true
}

// Helper функции для создания RPG персонажа

// CreateRPGCharacter создает сущность с полным набором RPG компонентов
func CreateRPGCharacter(world *ecs.World, level int) ecs.EntityID {
	entity := world.CreateEntity()

	// Базовые характеристики
	stats := &StatsComponent{
		Level:                 level,
		Strength:              10,
		Agility:               10,
		Intelligence:          10,
		Vitality:              10,
		Luck:                  5,
		Experience:            0,
		ExperienceToNextLevel: 100,
	}

	// Здоровье
	maxHP := float32(100 + stats.Vitality*10)
	health := &HealthComponent{
		Current:      maxHP,
		Max:          maxHP,
		Regeneration: 5.0,
	}

	// Мана
	maxMana := float32(50 + stats.Intelligence*5)
	mana := &ManaComponent{
		Current:      maxMana,
		Max:          maxMana,
		Regeneration: 3.0,
	}

	// Выносливость
	maxStamina := float32(100 + stats.Agility*8)
	stamina := &StaminaComponent{
		Current:      maxStamina,
		Max:          maxStamina,
		Regeneration: 10.0,
	}

	// Инвентарь
	inventory := &InventoryComponent{
		Slots:       make([]ItemSlot, 0),
		MaxSlots:    30,
		Gold:        0,
		MaxWeight:   100.0,
		CurrentWeight: 0.0,
	}

	// Экипировка
	equipment := &EquipmentComponent{
		Slots: make(map[EquipmentSlot]string),
	}

	// Журнал квестов
	questLog := &QuestLogComponent{
		ActiveQuests:    make([]Quest, 0),
		CompletedQuests: make([]string, 0),
		FailedQuests:    make([]string, 0),
	}

	// Добавляем компоненты
	world.AddComponent(entity, stats)
	world.AddComponent(entity, health)
	world.AddComponent(entity, mana)
	world.AddComponent(entity, stamina)
	world.AddComponent(entity, inventory)
	world.AddComponent(entity, equipment)
	world.AddComponent(entity, questLog)

	return entity
}
