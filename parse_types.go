package dodumap

type MappedMultilangCondition struct {
	Element   string            `json:"element"`
	ElementId int               `json:"element_id"`
	Operator  string            `json:"operator"`
	Value     int               `json:"value"`
	Templated map[string]string `json:"templated"`
}

// NodeType defines the type of a node - either an operator or an operand
type NodeType int

const (
	Operand  NodeType = iota // like "a" in "a & b"
	Operator                 // like "&" in "a & b"
)

type ConditionTreeNode struct {
	Value    string
	Type     NodeType
	Children []*ConditionTreeNode
}

type ConditionTreeNodeMapped struct {
	Value     *MappedMultilangCondition  `json:"value"`
	IsOperand bool                       `json:"is_operand"`
	Relation  *string                    `json:"relation"` // "and" or "or"
	Children  []*ConditionTreeNodeMapped `json:"children"`
}

type MappedMultilangRecipe struct {
	ResultId int                          `json:"result_id"`
	Entries  []MappedMultilangRecipeEntry `json:"entries"`
}

type MappedMultilangRecipeEntry struct {
	ItemId   int `json:"item_id"`
	Quantity int `json:"quantity"`
	//ItemType map[string]string `json:"item_type"`
}

type MappedMultilangSetReverseLink struct {
	Id   int               `json:"id"`
	Name map[string]string `json:"name"`
}

type MappedMultilangSet struct {
	AnkamaId   int                          `json:"ankama_id"`
	Name       map[string]string            `json:"name"`
	ItemIds    []int                        `json:"items"`
	Effects    [][]MappedMultilangSetEffect `json:"effects"`
	Level      int                          `json:"level"`
	IsCosmetic bool                         `json:"is_cosmetic"`
}

type MappedMultilangMount struct {
	AnkamaId   int                     `json:"ankama_id"`
	Name       map[string]string       `json:"name"`
	FamilyId   int                     `json:"family_id"`
	FamilyName map[string]string       `json:"family_name"`
	Effects    []MappedMultilangEffect `json:"effects"`
}

type MappedMultilangCharacteristic struct {
	Value map[string]string `json:"value"`
	Name  map[string]string `json:"name"`
}

type MappedMultilangSetEffect struct {
	Min              int               `json:"min"`
	Max              int               `json:"max"`
	Type             map[string]string `json:"type"`
	MinMaxIrrelevant int               `json:"min_max_irrelevant"`
	Templated        map[string]string `json:"templated"`
	ElementId        int               `json:"element_id"`
	IsMeta           bool              `json:"is_meta"`
	Active           bool              `json:"active"`
	ItemCombination  uint              `json:"item_combination"`
}

type MappedMultilangEffect struct {
	Min              int               `json:"min"`
	Max              int               `json:"max"`
	Type             map[string]string `json:"type"`
	MinMaxIrrelevant int               `json:"min_max_irrelevant"`
	Templated        map[string]string `json:"templated"`
	ElementId        int               `json:"element_id"`
	IsMeta           bool              `json:"is_meta"`
	Active           bool              `json:"active"`
}

type MappedMultilangItemType struct {
	Id          int               `json:"id"`
	Name        map[string]string `json:"name"`
	ItemTypeId  int               `json:"itemTypeId"`
	SuperTypeId int               `json:"superTypeId"`
	CategoryId  int               `json:"categoryId"`
}

type MappedMultilangItem struct {
	AnkamaId               int                             `json:"ankama_id"`
	Type                   MappedMultilangItemType         `json:"type"`
	Description            map[string]string               `json:"description"`
	Name                   map[string]string               `json:"name"`
	Image                  string                          `json:"image"`
	Conditions             []MappedMultilangCondition      `json:"conditions"`
	ConditionTree          *ConditionTreeNodeMapped        `json:"condition_tree"`
	Level                  int                             `json:"level"`
	UsedInRecipes          []int                           `json:"used_in_recipes"`
	Characteristics        []MappedMultilangCharacteristic `json:"characteristics"`
	Effects                []MappedMultilangEffect         `json:"effects"`
	DropMonsterIds         []int                           `json:"dropMonsterIds"`
	CriticalHitBonus       int                             `json:"criticalHitBonus"`
	TwoHanded              bool                            `json:"twoHanded"`
	MaxCastPerTurn         int                             `json:"maxCastPerTurn"`
	ApCost                 int                             `json:"apCost"`
	Range                  int                             `json:"range"`
	MinRange               int                             `json:"minRange"`
	CriticalHitProbability int                             `json:"criticalHitProbability"`
	Pods                   int                             `json:"pods"`
	IconId                 int                             `json:"iconId"`
	ParentSet              MappedMultilangSetReverseLink   `json:"parentSet"`
	HasParentSet           bool                            `json:"hasParentSet"`
}

type MappedMultilangNPCAlmanax struct {
	OfferingReceiver string   `json:"offeringReceiver"`
	Days             []string `json:"days"`
	Offering         struct {
		ItemId    int               `json:"itemId"`
		ItemName  map[string]string `json:"itemName"`
		Quantity  int               `json:"quantity"`
		ImageUrls struct {
			HD   string `json:"hd"`
			HQ   string `json:"hq"`
			SD   string `json:"sd"`
			Icon string `json:"icon"`
		}
	}
	Bonus       map[string]string `json:"bonus"`
	BonusType   map[string]string `json:"bonusType"`
	RewardKamas int               `json:"rewardKamas"`
}

type JSONGameSpellType struct {
	Id          int `json:"id"`
	LongNameId  int `json:"longNameId"`
	ShortNameId int `json:"shortNameId"`
}

func (i JSONGameSpellType) GetID() int {
	return i.Id
}

type JSONGameSpell struct {
	Id            int   `json:"id"`
	NameId        int   `json:"nameId"`
	DescriptionId int   `json:"descriptionId"`
	TypeId        int   `json:"typeId"`
	Order         int   `json:"order"`
	IconId        int   `json:"iconId"`
	SpellLevels   []int `json:"spellLevels"`
}

func (i JSONGameSpell) GetID() int {
	return i.Id
}

type JSONLangDict struct {
	Texts    map[string]string `json:"texts"`    // "1": "Account- oder Abohandel",
	IdText   map[string]int    `json:"idText"`   // "790745": 27679,
	NameText map[string]int    `json:"nameText"` // "ui.chat.check0": 65984
}

type JSONGameRecipe struct {
	Id            int   `json:"resultId"`
	NameId        int   `json:"resultNameId"`
	TypeId        int   `json:"resultTypeId"`
	Level         int   `json:"resultLevel"`
	IngredientIds []int `json:"ingredientIds"`
	Quantities    []int `json:"quantities"`
	JobId         int   `json:"jobId"`
	SkillId       int   `json:"skillId"`
}

func (i JSONGameRecipe) GetID() int {
	return i.Id
}

type LangDict struct {
	Texts    map[int]string
	IdText   map[int]int
	NameText map[string]int
}

type JSONGameBonus struct {
	Amount        int   `json:"amount"`
	Id            int   `json:"id"`
	CriterionsIds []int `json:"criterionsIds"`
	Type          int   `json:"type"`
}

func (i JSONGameBonus) GetID() int {
	return i.Id
}

type JSONGameAreaBounds struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

type JSONGameArea struct {
	Id              int                `json:"id"`
	NameId          int                `json:"nameId"`
	SuperAreaId     int                `json:"superAreaId"`
	ContainHouses   bool               `json:"containHouses"`
	ContainPaddocks bool               `json:"containPaddocks"`
	Bounds          JSONGameAreaBounds `json:"bounds"`
	WorldmapId      int                `json:"worldmapId"`
	HasWorldMap     bool               `json:"hasWorldMap"`
}

func (i JSONGameArea) GetID() int {
	return i.Id
}

type JSONGameItemPossibleEffect struct {
	EffectId     int `json:"effectId"`
	MinimumValue int `json:"diceNum"`
	MaximumValue int `json:"diceSide"`
	Value        int `json:"value"`

	BaseEffectId  int `json:"baseEffectId"`
	EffectElement int `json:"effectElement"`
	Dispellable   int `json:"dispellable"`
	SpellId       int `json:"spellId"`
	Duration      int `json:"duration"`
}

func (i JSONGameItemPossibleEffect) GetID() int {
	return i.EffectId
}

type JSONGameSet struct {
	Id      int                             `json:"id"`
	ItemIds []int                           `json:"items"`
	NameId  int                             `json:"nameId"`
	Effects [][]*JSONGameItemPossibleEffect `json:"effects"`
}

func (i JSONGameSet) GetID() int {
	return i.Id
}

type JSONGameItemType struct {
	Id          int `json:"id"`
	NameId      int `json:"nameId"`
	SuperTypeId int `json:"superTypeId"`
	CategoryId  int `json:"categoryId"`
}

func (i JSONGameItemType) GetID() int {
	return i.Id
}

type JSONGameEffect struct {
	Id                       int  `json:"id"`
	DescriptionId            int  `json:"descriptionId"`
	IconId                   int  `json:"iconId"`
	Characteristic           int  `json:"characteristic"`
	Category                 int  `json:"category"`
	UseDice                  bool `json:"useDice"`
	Active                   bool `json:"active"`
	TheoreticalDescriptionId int  `json:"theoreticalDescriptionId"`
	BonusType                int  `json:"bonusType"` // -1,0,+1
	ElementId                int  `json:"elementId"`
	UseInFight               bool `json:"useInFight"`
}

func (i JSONGameEffect) GetID() int {
	return i.Id
}

type JSONGameItem struct {
	Id            int `json:"id"`
	TypeId        int `json:"typeId"`
	DescriptionId int `json:"descriptionId"`
	IconId        int `json:"iconId"`
	NameId        int `json:"nameId"`
	Level         int `json:"level"`

	PossibleEffects        []*JSONGameItemPossibleEffect `json:"possibleEffects"`
	RecipeIds              []int                         `json:"recipeIds"`
	Pods                   int                           `json:"realWeight"`
	ParseEffects           bool                          `json:"useDice"`
	EvolutiveEffectIds     []int                         `json:"evolutiveEffectIds"`
	DropMonsterIds         []int                         `json:"dropMonsterIds"`
	ItemSetId              int                           `json:"itemSetId"`
	Criteria               string                        `json:"criteria"`
	CriticalHitBonus       int                           `json:"criticalHitBonus"`
	TwoHanded              bool                          `json:"twoHanded"`
	MaxCastPerTurn         int                           `json:"maxCastPerTurn"`
	ApCost                 int                           `json:"apCost"`
	Range                  int                           `json:"range"`
	MinRange               int                           `json:"minRange"`
	CriticalHitProbability int                           `json:"criticalHitProbability"`
}

func (i JSONGameItem) GetID() int {
	return i.Id
}

type JSONGameBreed struct {
	Id            int `json:"id"`
	ShortNameId   int `json:"shortNameId"`
	LongNameId    int `json:"longNameId"`
	DescriptionId int `json:"descriptionId"`
}

func (i JSONGameBreed) GetID() int {
	return i.Id
}

type JSONGameMount struct {
	Id       int                           `json:"id"`
	FamilyId int                           `json:"familyId"`
	NameId   int                           `json:"nameId"`
	Effects  []*JSONGameItemPossibleEffect `json:"effects"`
}

func (i JSONGameMount) GetID() int {
	return i.Id
}

type JSONGameMountFamily struct {
	Id      int    `json:"id"`
	NameId  int    `json:"nameId"`
	HeadUri string `json:"headUri"`
}

func (i JSONGameMountFamily) GetID() int {
	return i.Id
}

type JSONGameNPC struct {
	Id             int     `json:"id"`
	NameId         int     `json:"nameId"`
	DialogMessages [][]int `json:"dialogMessages"`
	DialogReplies  [][]int `json:"dialogReplies"`
	Actions        []int   `json:"actions"`
}

func (i JSONGameNPC) GetID() int {
	return i.Id
}

type JSONGameTitle struct {
	Id           int  `json:"id"`
	NameMaleId   int  `json:"nameMaleId"`
	NameFemaleId int  `json:"nameFemaleId"`
	Visible      bool `json:"visible"`
	CategoryId   int  `json:"categoryId"`
}

func (i JSONGameTitle) GetID() int {
	return i.Id
}

type JSONGameCoordinate struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type JSONGameQuestParameter struct {
	DungeonOnly bool `json:"dungeonOnly"`
	NumParams   int  `json:"numParams"`
	Parameter0  int  `json:"parameter0"`
	Parameter1  int  `json:"parameter1"`
	Parameter2  int  `json:"parameter2"`
	Parameter3  int  `json:"parameter3"`
	Parameter4  int  `json:"parameter4"`
}

type JSONGameQuestObjective struct {
	Id         int                    `json:"id"`
	Coords     JSONGameCoordinate     `json:"coords"`
	MapId      int                    `json:"mapId"`
	Parameters JSONGameQuestParameter `json:"parameters"`
	StepId     int                    `json:"stepId"`
	TypeId     int                    `json:"typeId"`
}

func (i JSONGameQuestObjective) GetID() int {
	return i.Id
}

type JSONGameQuestCategory struct {
	Id       int   `json:"id"`
	NameId   int   `json:"nameId"`
	Order    int   `json:"order"`
	QuestIds []int `json:"questIds"`
}

func (i JSONGameQuestCategory) GetID() int {
	return i.Id
}

type JSONGameQuest struct {
	Id             int    `json:"id"`
	NameId         int    `json:"nameId"`
	StepIds        []int  `json:"stepIds"`
	CategoryId     int    `json:"categoryId"`
	RepeatType     int    `json:"repeatType"`
	RepeatLimit    int    `json:"repeatLimit"`
	IsDungeonQuest bool   `json:"isDungeonQuest"`
	LevelMin       int    `json:"levelMin"`
	LevelMax       int    `json:"levelMax"`
	Followable     bool   `json:"followable"`
	IsPartyQuest   bool   `json:"isPartyQuest"`
	StartCriterion string `json:"startCriterion"`
}

func (i JSONGameQuest) GetID() int {
	return i.Id
}

type JSONGameQuestStepRewards struct {
	Id                        int     `json:"id"`
	ExperienceRatio           float64 `json:"experienceRatio"`
	KamasRatio                float64 `json:"kamasRatio"`
	ItemsReward               [][]int `json:"itemsReward"`
	KamasScaleWithPlayerLevel bool    `json:"kamasScaleWithPlayerLevel"`
	LevelMax                  int     `json:"levelMax"`
	LevelMin                  int     `json:"levelMin"`
	//SpellsReward []int `json:"spellsReward"`
	//EmotesReward []int `json:"emotesReward"`
	//TitlesReward []int `json:"titlesReward"`
	StepId int `json:"stepId"`
}

func (i JSONGameQuestStepRewards) GetID() int {
	return i.Id
}

type JSONGameAlamanaxCalendar struct {
	Id         int   `json:"id"`
	DescId     int   `json:"descId"`
	NameId     int   `json:"nameId"`
	NpcId      int   `json:"npcId"`
	BonusesIds []int `json:"bonusesIds"`
}

func (i JSONGameAlamanaxCalendar) GetID() int {
	return i.Id
}

type JSONGameQuestStep struct {
	Id            int     `json:"id"`
	DescriptionId int     `json:"descriptionId"`
	DialogId      int     `json:"dialogId"`
	NameId        int     `json:"nameId"`
	OptimalLevel  int     `json:"optimalLevel"`
	Duration      float64 `json:"duration"`
	ObjectiveIds  []int   `json:"objectiveIds"`
	RewardsIds    []int   `json:"rewardsIds"`
	QuestId       int     `json:"questId"`
}

func (i JSONGameQuestStep) GetID() int {
	return i.Id
}

type JSONGameData struct {
	Items            map[int]JSONGameItem
	Sets             map[int]JSONGameSet
	ItemTypes        map[int]JSONGameItemType
	effects          map[int]JSONGameEffect
	bonuses          map[int]JSONGameBonus
	Recipes          map[int]JSONGameRecipe
	spells           map[int]JSONGameSpell
	spellTypes       map[int]JSONGameSpellType
	areas            map[int]JSONGameArea
	Mounts           map[int]JSONGameMount
	classes          map[int]JSONGameBreed
	MountFamilys     map[int]JSONGameMountFamily
	npcs             map[int]JSONGameNPC
	titles           map[int]JSONGameTitle
	questSteps       map[int]JSONGameQuestStep
	questObjectives  map[int]JSONGameQuestObjective
	questCategories  map[int]JSONGameQuestCategory
	quests           map[int]JSONGameQuest
	questStepRewards map[int]JSONGameQuestStepRewards
	almanaxCalendars map[int]JSONGameAlamanaxCalendar
}
