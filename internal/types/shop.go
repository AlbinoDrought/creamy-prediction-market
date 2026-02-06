package types

// ShopItemCategory identifies the category an item belongs to
type ShopItemCategory string

const (
	ShopItemCategoryAvatarColor  = ShopItemCategory("avatar_color")
	ShopItemCategoryAvatarEmoji  = ShopItemCategory("avatar_emoji")
	ShopItemCategoryNameEmoji    = ShopItemCategory("name_emoji")
	ShopItemCategoryAvatarEffect = ShopItemCategory("avatar_effect")
	ShopItemCategoryNameEffect   = ShopItemCategory("name_effect")
	ShopItemCategoryNameBold     = ShopItemCategory("name_bold")
	ShopItemCategoryNameFont     = ShopItemCategory("name_font")
	ShopItemCategoryTitle        = ShopItemCategory("title")
	ShopItemCategoryHat          = ShopItemCategory("hat")
	ShopItemCategoryAvatarItem   = ShopItemCategory("avatar_item")
	ShopItemCategoryGlobalAction = ShopItemCategory("global_action")
)

// ShopItem represents a purchasable cosmetic item
type ShopItem struct {
	ID       string           `json:"id"`
	Category ShopItemCategory `json:"category"`
	Name     string           `json:"name"`
	Price    int64            `json:"price"`
	Icon     string           `json:"icon"`
	// Value is the resolved cosmetic data (hex colors, emoji, CSS class name, title text, etc.)
	Value string `json:"value"`
	// Consumable items are used immediately and not added to inventory
	Consumable bool `json:"consumable"`
}

// AllShopItems is the complete catalog of purchasable items
var AllShopItems = []ShopItem{
	// Avatar Colors
	{ID: "avatar_color_sunset", Category: ShopItemCategoryAvatarColor, Name: "Sunset", Price: 5, Icon: "🌅", Value: "#F97316,#EF4444"},
	{ID: "avatar_color_ocean", Category: ShopItemCategoryAvatarColor, Name: "Ocean", Price: 5, Icon: "🌊", Value: "#3B82F6,#06B6D4"},
	{ID: "avatar_color_forest", Category: ShopItemCategoryAvatarColor, Name: "Forest", Price: 5, Icon: "🌲", Value: "#22C55E,#10B981"},
	{ID: "avatar_color_neon", Category: ShopItemCategoryAvatarColor, Name: "Neon", Price: 10, Icon: "💜", Value: "#EC4899,#06B6D4"},
	{ID: "avatar_color_fire", Category: ShopItemCategoryAvatarColor, Name: "Fire", Price: 10, Icon: "🔥", Value: "#EF4444,#F59E0B"},
	{ID: "avatar_color_ice", Category: ShopItemCategoryAvatarColor, Name: "Ice", Price: 10, Icon: "🧊", Value: "#93C5FD,#E0F2FE"},
	{ID: "avatar_color_slime", Category: ShopItemCategoryAvatarColor, Name: "Slime", Price: 10, Icon: "🦠", Value: "#CCFF00,#69B204"},
	{ID: "avatar_color_rose", Category: ShopItemCategoryAvatarColor, Name: "Rose", Price: 10, Icon: "🌹", Value: "#E11D48,#BE123C"},
	{ID: "avatar_color_steel", Category: ShopItemCategoryAvatarColor, Name: "Steel", Price: 15, Icon: "🤖", Value: "#4B5563,#9CA3AF"},
	{ID: "avatar_color_royal", Category: ShopItemCategoryAvatarColor, Name: "Royal", Price: 15, Icon: "👑", Value: "#7C3AED,#FACC15"},
	{ID: "avatar_color_midnight", Category: ShopItemCategoryAvatarColor, Name: "Midnight", Price: 15, Icon: "🌙", Value: "#1E1B4B,#7C3AED"},
	{ID: "avatar_color_sunshine", Category: ShopItemCategoryAvatarColor, Name: "Sunshine", Price: 15, Icon: "🌞", Value: "#F7931A,#FFD700"},
	{ID: "avatar_color_matrix", Category: ShopItemCategoryAvatarColor, Name: "Matrix", Price: 15, Icon: "🟢", Value: "#000000,#00FF00"},
	{ID: "avatar_color_gamer", Category: ShopItemCategoryAvatarColor, Name: "Gamer", Price: 20, Icon: "🎮", Value: "#9D00FF,#00F0FF"},
	{ID: "avatar_color_gold", Category: ShopItemCategoryAvatarColor, Name: "Gold", Price: 25, Icon: "🥇", Value: "#BF953F,#FCF6BA"},
	{ID: "avatar_color_void", Category: ShopItemCategoryAvatarColor, Name: "Void", Price: 30, Icon: "🕳️", Value: "#000000,#312E81"},

	// Avatar Emojis
	{ID: "avatar_emoji_cool", Category: ShopItemCategoryAvatarEmoji, Name: "Cool", Price: 5, Icon: "😎", Value: "😎"},
	{ID: "avatar_emoji_cowboy", Category: ShopItemCategoryAvatarEmoji, Name: "Cowboy", Price: 5, Icon: "🤠", Value: "🤠"},
	// {ID: "avatar_emoji_sad", Category: ShopItemCategoryAvatarEmoji, Name: "Sad", Price: 5, Icon: "😭", Value: "😭"}, // sample: give this when user loses 100 tokens in a bet
	{ID: "avatar_emoji_ghost", Category: ShopItemCategoryAvatarEmoji, Name: "Ghost", Price: 5, Icon: "👻", Value: "👻"},
	{ID: "avatar_emoji_robot", Category: ShopItemCategoryAvatarEmoji, Name: "Robot", Price: 5, Icon: "🤖", Value: "🤖"},
	{ID: "avatar_emoji_fox", Category: ShopItemCategoryAvatarEmoji, Name: "Fox", Price: 10, Icon: "🦊", Value: "🦊"},
	{ID: "avatar_emoji_gorilla", Category: ShopItemCategoryAvatarEmoji, Name: "Gorilla", Price: 10, Icon: "🦍", Value: "🦍"},
	{ID: "avatar_emoji_monkey", Category: ShopItemCategoryAvatarEmoji, Name: "Monkey", Price: 10, Icon: "🐒", Value: "🐒"},
	{ID: "avatar_emoji_orangutan", Category: ShopItemCategoryAvatarEmoji, Name: "Orangutan", Price: 10, Icon: "🦧", Value: "🦧"},
	{ID: "avatar_emoji_peach", Category: ShopItemCategoryAvatarEmoji, Name: "Peach", Price: 15, Icon: "🍑", Value: "🍑"},
	{ID: "avatar_emoji_burrito", Category: ShopItemCategoryAvatarEmoji, Name: "Burrito", Price: 15, Icon: "🌯", Value: "🌯"},
	{ID: "avatar_emoji_hoddog", Category: ShopItemCategoryAvatarEmoji, Name: "Hoddog", Price: 15, Icon: "🌭", Value: "🌭"}, // spelling is intentional
	{ID: "avatar_emoji_dragon", Category: ShopItemCategoryAvatarEmoji, Name: "Dragon", Price: 15, Icon: "🐉", Value: "🐉"},
	{ID: "avatar_emoji_money_face", Category: ShopItemCategoryAvatarEmoji, Name: "Money Face", Price: 15, Icon: "🤑", Value: "🤑"},
	{ID: "avatar_emoji_crown", Category: ShopItemCategoryAvatarEmoji, Name: "Crown", Price: 20, Icon: "👑", Value: "👑"},
	{ID: "avatar_emoji_crystal", Category: ShopItemCategoryAvatarEmoji, Name: "Crystal Ball", Price: 20, Icon: "🔮", Value: "🔮"},
	{ID: "avatar_emoji_sparkles", Category: ShopItemCategoryAvatarEmoji, Name: "Sparkles", Price: 25, Icon: "✨", Value: "✨"},
	{ID: "avatar_emoji_star", Category: ShopItemCategoryAvatarEmoji, Name: "Star", Price: 25, Icon: "🌟", Value: "🌟"},

	// Name Emojis
	{ID: "name_emoji_canada", Category: ShopItemCategoryNameEmoji, Name: "Canada", Price: 0, Icon: "🇨🇦", Value: "🇨🇦"},
	{ID: "name_emoji_star", Category: ShopItemCategoryNameEmoji, Name: "Star", Price: 5, Icon: "⭐", Value: "⭐"},
	{ID: "name_emoji_fire", Category: ShopItemCategoryNameEmoji, Name: "Fire", Price: 5, Icon: "🔥", Value: "🔥"},
	{ID: "name_emoji_cheers", Category: ShopItemCategoryNameEmoji, Name: "Cheers", Price: 5, Icon: "🍻", Value: "🍻"},
	{ID: "name_emoji_skull", Category: ShopItemCategoryNameEmoji, Name: "Skull", Price: 10, Icon: "💀", Value: "💀"},
	{ID: "name_emoji_gem", Category: ShopItemCategoryNameEmoji, Name: "Gem", Price: 10, Icon: "💎", Value: "💎"},
	{ID: "name_emoji_lightning", Category: ShopItemCategoryNameEmoji, Name: "Lightning", Price: 10, Icon: "⚡", Value: "⚡"},
	{ID: "name_emoji_crown", Category: ShopItemCategoryNameEmoji, Name: "Crown", Price: 15, Icon: "👑", Value: "👑"},
	{ID: "name_emoji_rainbow", Category: ShopItemCategoryNameEmoji, Name: "Rainbow", Price: 15, Icon: "🌈", Value: "🌈"},
	{ID: "name_emoji_sparkles", Category: ShopItemCategoryNameEmoji, Name: "Sparkles", Price: 20, Icon: "✨", Value: "✨"},
	{ID: "name_emoji_rocket", Category: ShopItemCategoryNameEmoji, Name: "Rocket", Price: 20, Icon: "🚀", Value: "🚀"},
	{ID: "name_emoji_trophy", Category: ShopItemCategoryNameEmoji, Name: "Trophy", Price: 25, Icon: "🏆", Value: "🏆"},
	{ID: "name_emoji_moneybags", Category: ShopItemCategoryNameEmoji, Name: "Moneybags", Price: 30, Icon: "💰", Value: "💰"},

	// Avatar Effects
	{ID: "avatar_effect_glow", Category: ShopItemCategoryAvatarEffect, Name: "Glow", Price: 10, Icon: "💡", Value: "glow"},
	{ID: "avatar_effect_sparkle", Category: ShopItemCategoryAvatarEffect, Name: "Sparkle", Price: 15, Icon: "✨", Value: "sparkle"},
	{ID: "avatar_effect_fire", Category: ShopItemCategoryAvatarEffect, Name: "Fire", Price: 20, Icon: "🔥", Value: "fire"},
	{ID: "avatar_effect_rainbow", Category: ShopItemCategoryAvatarEffect, Name: "Rainbow", Price: 25, Icon: "🌈", Value: "rainbow"},

	// Name Effects
	{ID: "name_effect_glow", Category: ShopItemCategoryNameEffect, Name: "Glow", Price: 10, Icon: "💡", Value: "glow"},
	{ID: "name_effect_sparkle", Category: ShopItemCategoryNameEffect, Name: "Sparkle", Price: 15, Icon: "✨", Value: "sparkle"},
	{ID: "name_effect_rainbow", Category: ShopItemCategoryNameEffect, Name: "Rainbow", Price: 20, Icon: "🌈", Value: "rainbow"},

	// Name Bold
	{ID: "name_bold", Category: ShopItemCategoryNameBold, Name: "Bold Name", Price: 5, Icon: "🅱️", Value: "bold"},

	// Name Fonts
	{ID: "name_font_serif", Category: ShopItemCategoryNameFont, Name: "Serif", Price: 5, Icon: "🔤", Value: "serif"},
	{ID: "name_font_mono", Category: ShopItemCategoryNameFont, Name: "Mono", Price: 5, Icon: "💻", Value: "mono"},
	{ID: "name_font_cursive", Category: ShopItemCategoryNameFont, Name: "Cursive", Price: 10, Icon: "✍️", Value: "cursive"},
	{ID: "name_font_comic", Category: ShopItemCategoryNameFont, Name: "Comic Sans", Price: 15, Icon: "🤪", Value: "comic"},

	// Titles - Betting
	{ID: "title_broke", Category: ShopItemCategoryTitle, Name: "Broke", Price: 1, Value: "Broke"},
	{ID: "title_gambler", Category: ShopItemCategoryTitle, Name: "Gambler", Price: 5, Value: "Gambler"},
	{ID: "title_all_in", Category: ShopItemCategoryTitle, Name: "All In", Price: 10, Value: "All In"},
	{ID: "title_high_roller", Category: ShopItemCategoryTitle, Name: "High Roller", Price: 10, Value: "High Roller"},
	{ID: "title_degen", Category: ShopItemCategoryTitle, Name: "Degen", Price: 10, Value: "Degen"},
	{ID: "title_oracle", Category: ShopItemCategoryTitle, Name: "Oracle", Price: 15, Value: "Oracle"},
	{ID: "title_whale", Category: ShopItemCategoryTitle, Name: "Whale", Price: 20, Value: "Whale"},
	{ID: "title_prophet", Category: ShopItemCategoryTitle, Name: "Prophet", Price: 25, Value: "Prophet"},

	// Titles - Tech
	{ID: "title_10x_engineer", Category: ShopItemCategoryTitle, Name: "10x Engineer", Price: 10, Value: "10x Engineer"},
	{ID: "title_senior_intern", Category: ShopItemCategoryTitle, Name: "Senior Intern", Price: 5, Value: "Senior Intern"},
	{ID: "title_stack_overflow", Category: ShopItemCategoryTitle, Name: "Stack Overflow Survivor", Price: 10, Value: "Stack Overflow Survivor"},
	{ID: "title_yaml_engineer", Category: ShopItemCategoryTitle, Name: "YAML Engineer", Price: 10, Value: "YAML Engineer"},
	{ID: "title_git_blame", Category: ShopItemCategoryTitle, Name: "git blame Enthusiast", Price: 10, Value: "git blame Enthusiast"},
	{ID: "title_lgtm", Category: ShopItemCategoryTitle, Name: "LGTM", Price: 5, Value: "LGTM"},
	{ID: "title_it_works_locally", Category: ShopItemCategoryTitle, Name: "Works On My Machine", Price: 10, Value: "Works On My Machine"},
	{ID: "title_refactorer", Category: ShopItemCategoryTitle, Name: "Compulsive Refactorer", Price: 15, Value: "Compulsive Refactorer"},
	{ID: "title_rubber_duck", Category: ShopItemCategoryTitle, Name: "Rubber Duck", Price: 5, Value: "Rubber Duck"},
	{ID: "title_prompt_engineer", Category: ShopItemCategoryTitle, Name: "Prompt Engineer", Price: 15, Value: "Prompt Engineer"},
	{ID: "title_undefined", Category: ShopItemCategoryTitle, Name: "undefined", Price: 10, Value: "undefined"},
	{ID: "title_404", Category: ShopItemCategoryTitle, Name: "404 Not Found", Price: 10, Value: "404 Not Found"},

	// Titles - Enterprise / Product
	{ID: "title_synergy", Category: ShopItemCategoryTitle, Name: "Synergy Specialist", Price: 10, Value: "Synergy Specialist"},
	{ID: "title_thought_leader", Category: ShopItemCategoryTitle, Name: "Thought Leader", Price: 15, Value: "Thought Leader"},
	{ID: "title_disruptor", Category: ShopItemCategoryTitle, Name: "Disruptor", Price: 15, Value: "Disruptor"},
	{ID: "title_stakeholder", Category: ShopItemCategoryTitle, Name: "Key Stakeholder", Price: 10, Value: "Key Stakeholder"},
	{ID: "title_blocker", Category: ShopItemCategoryTitle, Name: "Blocker", Price: 5, Value: "Blocker"},
	{ID: "title_circle_back", Category: ShopItemCategoryTitle, Name: "Let's Circle Back", Price: 10, Value: "Let's Circle Back"},
	{ID: "title_per_my_email", Category: ShopItemCategoryTitle, Name: "Per My Last Email", Price: 15, Value: "Per My Last Email"},
	{ID: "title_vc_funded", Category: ShopItemCategoryTitle, Name: "VC Funded", Price: 20, Value: "VC Funded"},

	// Titles - Drinking
	{ID: "title_drunk", Category: ShopItemCategoryTitle, Name: "Drunk", Price: 5, Value: "Drunk"},
	{ID: "title_sommelier", Category: ShopItemCategoryTitle, Name: "Sommelier", Price: 15, Value: "Sommelier"},
	{ID: "title_designated_drinker", Category: ShopItemCategoryTitle, Name: "Designated Drinker", Price: 10, Value: "Designated Drinker"},
	{ID: "title_on_the_rocks", Category: ShopItemCategoryTitle, Name: "On The Rocks", Price: 10, Value: "On The Rocks"},
	{ID: "title_bottomless_mimosas", Category: ShopItemCategoryTitle, Name: "Bottomless Mimosas", Price: 15, Value: "Bottomless Mimosas"},
	{ID: "title_happy_hour", Category: ShopItemCategoryTitle, Name: "Happy Hour", Price: 5, Value: "Happy Hour"},
	{ID: "title_liquid_courage", Category: ShopItemCategoryTitle, Name: "Liquid Courage", Price: 10, Value: "Liquid Courage"},
	{ID: "title_beer_goggles", Category: ShopItemCategoryTitle, Name: "Beer Goggles", Price: 10, Value: "Beer Goggles"},

	// Titles - Goofy
	{ID: "title_gremlin", Category: ShopItemCategoryTitle, Name: "Gremlin", Price: 5, Value: "Gremlin"},
	{ID: "title_chaos_agent", Category: ShopItemCategoryTitle, Name: "Chaos Agent", Price: 15, Value: "Chaos Agent"},
	{ID: "title_npc", Category: ShopItemCategoryTitle, Name: "NPC", Price: 5, Value: "NPC"},
	{ID: "title_main_character", Category: ShopItemCategoryTitle, Name: "Main Character", Price: 20, Value: "Main Character"},
	{ID: "title_touch_grass", Category: ShopItemCategoryTitle, Name: "Needs To Touch Grass", Price: 10, Value: "Needs To Touch Grass"},
	{ID: "title_lurker", Category: ShopItemCategoryTitle, Name: "Lurker", Price: 5, Value: "Lurker"},

	// Hats
	{ID: "hat_crown", Category: ShopItemCategoryHat, Name: "Crown", Price: 20, Icon: "👑", Value: "👑"},
	{ID: "hat_tophat", Category: ShopItemCategoryHat, Name: "Top Hat", Price: 15, Icon: "🎩", Value: "🎩"},
	{ID: "hat_siren", Category: ShopItemCategoryHat, Name: "Siren", Price: 10, Icon: "🚨", Value: "🚨"},
	{ID: "hat_bow", Category: ShopItemCategoryHat, Name: "Bow", Price: 10, Icon: "🎀", Value: "🎀"},
	{ID: "hat_thinker", Category: ShopItemCategoryHat, Name: "Thinker", Price: 10, Icon: "💡", Value: "💡"},
	// {ID: "hat_sun", Category: ShopItemCategoryHat, Name: "Sunhat", Price: 10, Icon: "👒", Value: "👒"}, // doesn't work
	{ID: "hat_cap", Category: ShopItemCategoryHat, Name: "Cap", Price: 5, Icon: "🧢", Value: "🧢"},

	// Avatar Items
	{ID: "avatar_item_surrender", Category: ShopItemCategoryAvatarItem, Name: "Surrender", Price: 1, Icon: "🏳️", Value: "🏳️"},
	{ID: "avatar_item_wine", Category: ShopItemCategoryAvatarItem, Name: "Wine", Price: 5, Icon: "🍷", Value: "🍷"},
	{ID: "avatar_item_beer", Category: ShopItemCategoryAvatarItem, Name: "Beer", Price: 5, Icon: "🍺", Value: "🍺"},
	{ID: "avatar_item_liquor", Category: ShopItemCategoryAvatarItem, Name: "Liquor", Price: 5, Icon: "🥃", Value: "🥃"},
	{ID: "avatar_item_boba", Category: ShopItemCategoryAvatarItem, Name: "Boba", Price: 5, Icon: "🧋", Value: "🧋"},
	{ID: "avatar_item_dice", Category: ShopItemCategoryAvatarItem, Name: "Dice", Price: 5, Icon: "🎲", Value: "🎲"},
	{ID: "avatar_item_football", Category: ShopItemCategoryAvatarItem, Name: "Football", Price: 5, Icon: "🏈", Value: "🏈"},
	{ID: "avatar_item_soccer", Category: ShopItemCategoryAvatarItem, Name: "Soccer", Price: 5, Icon: "⚽", Value: "⚽"},
	{ID: "avatar_item_basketball", Category: ShopItemCategoryAvatarItem, Name: "Basketball", Price: 5, Icon: "🏀", Value: "🏀"},
	{ID: "avatar_item_volleyball", Category: ShopItemCategoryAvatarItem, Name: "Volleyball", Price: 5, Icon: "🏐", Value: "🏐"},
	{ID: "avatar_item_baseball", Category: ShopItemCategoryAvatarItem, Name: "Baseball", Price: 5, Icon: "⚾", Value: "⚾"},
	{ID: "avatar_item_tennis", Category: ShopItemCategoryAvatarItem, Name: "Tennis", Price: 10, Icon: "🎾", Value: "🎾"},
	{ID: "avatar_item_magic_wand", Category: ShopItemCategoryAvatarItem, Name: "Magic Wand", Price: 10, Icon: "🪄", Value: "🪄"},
	{ID: "avatar_item_banjo", Category: ShopItemCategoryAvatarItem, Name: "Banjo", Price: 10, Icon: "🪕", Value: "🪕"},
	{ID: "avatar_item_guitar", Category: ShopItemCategoryAvatarItem, Name: "Guitar", Price: 10, Icon: "🎸", Value: "🎸"},
	{ID: "avatar_item_rose", Category: ShopItemCategoryAvatarItem, Name: "Rose", Price: 10, Icon: "🌹", Value: "🌹"},
	{ID: "avatar_item_cookie", Category: ShopItemCategoryAvatarItem, Name: "Cookie", Price: 5, Icon: "🍪", Value: "🍪"},
	{ID: "avatar_item_pizza", Category: ShopItemCategoryAvatarItem, Name: "Pizza", Price: 10, Icon: "🍕", Value: "🍕"},
	{ID: "avatar_item_burrito", Category: ShopItemCategoryAvatarItem, Name: "Burrito", Price: 10, Icon: "🌯", Value: "🌯"},
	{ID: "avatar_item_happycat", Category: ShopItemCategoryAvatarItem, Name: "Happy Cat", Price: 10, Icon: "😺", Value: "😺"},
	{ID: "avatar_item_sadcat", Category: ShopItemCategoryAvatarItem, Name: "Sad Cat", Price: 10, Icon: "😿", Value: "😿"},
	{ID: "avatar_item_dog", Category: ShopItemCategoryAvatarItem, Name: "Dog", Price: 10, Icon: "🐕", Value: "🐕"},
	{ID: "avatar_item_zebra", Category: ShopItemCategoryAvatarItem, Name: "Zebra", Price: 15, Icon: "🦓", Value: "🦓"},
	{ID: "avatar_item_orangutan", Category: ShopItemCategoryAvatarItem, Name: "Orangutan", Price: 15, Icon: "🦧", Value: "🦧"},
	{ID: "avatar_item_peace", Category: ShopItemCategoryAvatarItem, Name: "Peace", Price: 5, Icon: "✌️", Value: "✌️"},
	{ID: "avatar_item_controller", Category: ShopItemCategoryAvatarItem, Name: "Controller", Price: 15, Icon: "🎮", Value: "🎮"},

	// Global Actions (consumable)
	{ID: "global_snowflakes", Category: ShopItemCategoryGlobalAction, Name: "Snowflakes", Price: 25, Icon: "❄️", Value: "snowflakes", Consumable: true},
	{ID: "global_fireworks", Category: ShopItemCategoryGlobalAction, Name: "Fireworks", Price: 25, Icon: "🎆", Value: "fireworks", Consumable: true},
}

func init() {
	seen := map[string]interface{}{}
	for _, v := range AllShopItems {
		if _, ok := seen[v.ID]; ok {
			panic(v.ID + " seen more than once!")
		}
		seen[v.ID] = true
	}
}

// GetShopItemByID returns a shop item by its ID
func GetShopItemByID(id string) (ShopItem, bool) {
	for _, item := range AllShopItems {
		if item.ID == id {
			return item, true
		}
	}
	return ShopItem{}, false
}
