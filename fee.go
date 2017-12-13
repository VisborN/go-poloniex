package poloniex

type Fees struct {
	MakerFee        string `json:"makerFee"`
	TakerFee        string `json:"takerFee"`
	ThirtyDayVolume string `json:"thirtyDayVolume"`
	NextTier        string `json:"nextTier"`
}
