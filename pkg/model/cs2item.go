package model

type Sticker struct {
	StickerID      int      `json:"stickerId,omitempty"`
	Slot           int      `json:"slot,omitempty"`
	Wear           float64  `json:"wear,omitempty"`
	Name           string   `json:"name,omitempty"`
	ImageURL       string   `json:"imageUrl,omitempty"`
	ReferencePrice float64  `json:"referencePrice,omitempty"`
	OffsetX        *float64 `json:"offsetX,omitempty"`
	OffsetY        *float64 `json:"offsetY,omitempty"`
}

type Keychain struct {
	Name     string `json:"name,omitempty"`
	ImageURL string `json:"imageUrl,omitempty"`
}

type CS2Item struct {
	AssetID        string `json:"assetId,omitempty"`
	ClassID        string `json:"classId,omitempty"`
	InstanceID     string `json:"instanceId,omitempty"`
	GoodsID        int    `json:"goodsId,omitempty"`
	CsqaqGoodsID   int    `json:"csqaqGoodsId,omitempty"`
	MarketHashName string `json:"marketHashName,omitempty"`
	ItemName       string `json:"itemName,omitempty"`
	IconURL        string `json:"iconUrl,omitempty"`

	PaintWear  float64 `json:"paintWear,omitempty"`
	PaintSeed  int     `json:"paintSeed,omitempty"`
	PaintIndex int     `json:"paintIndex,omitempty"`

	CategoryGroup string `json:"categoryGroup,omitempty"`
	Exterior      string `json:"exterior,omitempty"`
	Rarity        string `json:"rarity,omitempty"`
	WeaponType    string `json:"weaponType,omitempty"`
	WeaponName    string `json:"weaponName,omitempty"`
	Quality       string `json:"quality,omitempty"`
	Series        string `json:"series,omitempty"`
	Itemset       string `json:"itemset,omitempty"`
	WeaponCase    string `json:"weaponCase,omitempty"`
	Custom        string `json:"custom,omitempty"`

	Stickers  []Sticker  `gorm:"serializer:json" json:"stickers,omitempty"`
	Keychains []Keychain `gorm:"serializer:json" json:"keychains,omitempty"`

	TradableUnfrozenTime *int64 `json:"tradableUnfrozenTime,omitempty"`
}
