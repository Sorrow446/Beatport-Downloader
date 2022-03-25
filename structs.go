package main

type Transport struct{}

type Config struct {
	Email         string
	Password      string
	Urls          []string
	OutPath       string
	AlbumTemplate string
	TrackTemplate string
	MaxCover      bool
	OmitOrigMix   bool
	KeepCover     bool
}

type Args struct {
	Urls          []string `arg:"positional, required"`
	OutPath       string   `arg:"-o" help:"Where to download to. Path will be made if it doesn't already exist."`
	MaxCover      bool     `arg:"-m" help:"true = max cover size, false = 600x600."`
	AlbumTemplate string   `arg:"-a" help:"Album folder naming template. Vars: album, albumArtist, catalogNumber, upc, year."`
	TrackTemplate string   `arg:"-t" help:"Track filename naming template. Vars: album, albumArtist, artist, bpm, genre, isrc, title, track, trackPad, trackTotal, year."`
}

type UserSub struct {
	Subscription struct {
		UpdatedPersonID       int         `json:"updated_person_id"`
		CreatedPersonID       int         `json:"created_person_id"`
		ID                    int         `json:"id"`
		Person                int         `json:"person"`
		Notes                 interface{} `json:"notes"`
		FreeTrialStartDate    string      `json:"free_trial_start_date"`
		FreeTrialEndDate      string      `json:"free_trial_end_date"`
		RecurlySubscriptionID string      `json:"recurly_subscription_id"`
		Bundle                struct {
			UpdatedPersonID int    `json:"updated_person_id"`
			CreatedPersonID int    `json:"created_person_id"`
			ID              int    `json:"id"`
			Name            string `json:"name"`
			Description     string `json:"description"`
			Enabled         bool   `json:"enabled"`
			PlanCode        string `json:"plan_code"`
		} `json:"bundle"`
		StartDate                 string `json:"start_date"`
		EndDate                   string `json:"end_date"`
		RecurlySubscriptionStatus string `json:"recurly_subscription_status"`
		RecurlyInvoiceStatus      string `json:"recurly_invoice_status"`
	} `json:"subscription"`
	BillingInfo struct {
		Type      string `json:"type"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Address1  string `json:"address1"`
		Address2  string `json:"address2"`
		City      string `json:"city"`
		State     string `json:"state"`
		Zip       string `json:"zip"`
		Country   string `json:"country"`
		CardType  string `json:"card_type"`
		LastFour  string `json:"last_four"`
	} `json:"billing_info"`
	Active bool     `json:"active"`
	Status []string `json:"status"`
}

type Artist struct {
	ID    int `json:"id"`
	Image struct {
		ID         int    `json:"id"`
		URI        string `json:"uri"`
		DynamicURI string `json:"dynamic_uri"`
	} `json:"image"`
	Name string `json:"name"`
	Slug string `json:"slug"`
	URL  string `json:"url"`
}

type AlbumMeta struct {
	Artists  []Artist `json:"artists"`
	BpmRange struct {
		Min int `json:"min"`
		Max int `json:"max"`
	} `json:"bpm_range"`
	CatalogNumber string      `json:"catalog_number"`
	Desc          string      `json:"desc"`
	Enabled       bool        `json:"enabled"`
	EncodedDate   string      `json:"encoded_date"`
	Exclusive     bool        `json:"exclusive"`
	Grid          interface{} `json:"grid"`
	ID            int         `json:"id"`
	Image         struct {
		ID         int    `json:"id"`
		URI        string `json:"uri"`
		DynamicURI string `json:"dynamic_uri"`
	} `json:"image"`
	IsAvailableForStreaming bool `json:"is_available_for_streaming"`
	Label                   struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Image struct {
			ID         int    `json:"id"`
			URI        string `json:"uri"`
			DynamicURI string `json:"dynamic_uri"`
		} `json:"image"`
		Slug string `json:"slug"`
	} `json:"label"`
	Name           string      `json:"name"`
	NewReleaseDate string      `json:"new_release_date"`
	OverridePrice  interface{} `json:"override_price"`
	PreOrder       bool        `json:"pre_order"`
	PreOrderDate   interface{} `json:"pre_order_date"`
	Price          struct {
		Code    string  `json:"code"`
		Symbol  string  `json:"symbol"`
		Value   float64 `json:"value"`
		Display string  `json:"display"`
	} `json:"price"`
	PriceOverrideFirm bool          `json:"price_override_firm"`
	PublishDate       string        `json:"publish_date"`
	Remixers          []interface{} `json:"remixers"`
	Slug              string        `json:"slug"`
	Tracks            []string      `json:"tracks"`
	TrackCount        int           `json:"track_count"`
	Type              struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"type"`
	Upc     interface{} `json:"upc"`
	Updated string      `json:"updated"`
	IsHype  bool        `json:"is_hype"`
}

type TrackMeta struct {
	Artists            []Artist    `json:"artists"`
	AudioFormat        interface{} `json:"audio_format"`
	AvailableWorldwide bool        `json:"available_worldwide"`
	Bpm                int         `json:"bpm"`
	CatalogNumber      string      `json:"catalog_number"`
	CurrentStatus      struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"current_status"`
	Desc            string `json:"desc"`
	Enabled         bool   `json:"enabled"`
	EncodeStatus    string `json:"encode_status"`
	EncodedDate     string `json:"encoded_date"`
	Exclusive       bool   `json:"exclusive"`
	ExclusivePeriod struct {
		Days        int    `json:"days"`
		Description string `json:"description"`
		ID          int    `json:"id"`
		URL         string `json:"url"`
	} `json:"exclusive_period"`
	FreeDownloads         []interface{} `json:"free_downloads"`
	FreeDownloadStartDate interface{}   `json:"free_download_start_date"`
	FreeDownloadEndDate   interface{}   `json:"free_download_end_date"`
	Genre                 struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
		URL  string `json:"url"`
	} `json:"genre"`
	Hidden bool `json:"hidden"`
	ID     int  `json:"id"`
	Image  struct {
		ID         int    `json:"id"`
		URI        string `json:"uri"`
		DynamicURI string `json:"dynamic_uri"`
	} `json:"image"`
	IsAvailableForStreaming bool        `json:"is_available_for_streaming"`
	IsClassic               bool        `json:"is_classic"`
	Isrc                    interface{} `json:"isrc"`
	Key                     struct {
		CamelotNumber int    `json:"camelot_number"`
		CamelotLetter string `json:"camelot_letter"`
		ChordType     struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"chord_type"`
		ID      int    `json:"id"`
		IsSharp bool   `json:"is_sharp"`
		IsFlat  bool   `json:"is_flat"`
		Letter  string `json:"letter"`
		Name    string `json:"name"`
		URL     string `json:"url"`
	} `json:"key"`
	LabelTrackIdentifier string      `json:"label_track_identifier"`
	Length               string      `json:"length"`
	LengthMs             int         `json:"length_ms"`
	MixName              string      `json:"mix_name"`
	Name                 string      `json:"name"`
	NewReleaseDate       string      `json:"new_release_date"`
	Number               int         `json:"number"`
	PreOrder             bool        `json:"pre_order"`
	PreOrderDate         interface{} `json:"pre_order_date"`
	Price                struct {
		Code    string  `json:"code"`
		Symbol  string  `json:"symbol"`
		Value   float64 `json:"value"`
		Display string  `json:"display"`
	} `json:"price"`
	PublishDate   string `json:"publish_date"`
	PublishStatus string `json:"publish_status"`
	Release       struct {
		ID    int    `json:"id"`
		Name  string `json:"name"`
		Image struct {
			ID         int    `json:"id"`
			URI        string `json:"uri"`
			DynamicURI string `json:"dynamic_uri"`
		} `json:"image"`
		Label struct {
			ID    int    `json:"id"`
			Name  string `json:"name"`
			Image struct {
				ID         int    `json:"id"`
				URI        string `json:"uri"`
				DynamicURI string `json:"dynamic_uri"`
			} `json:"image"`
			Slug string `json:"slug"`
		} `json:"label"`
		Slug string `json:"slug"`
	} `json:"release"`
	Remixers []interface{} `json:"remixers"`
	SaleType struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"sale_type"`
	SampleURL        string      `json:"sample_url"`
	SampleStartMs    int         `json:"sample_start_ms"`
	SampleEndMs      int         `json:"sample_end_ms"`
	Slug             string      `json:"slug"`
	SubGenre         interface{} `json:"sub_genre"`
	WasEverExclusive bool        `json:"was_ever_exclusive"`
	IsHype           bool        `json:"is_hype"`
}

type TrackStream struct {
	StreamURL     string `json:"stream_url"`
	SampleStartMs int    `json:"sample_start_ms"`
	SampleEndMs   int    `json:"sample_end_ms"`
}

type Segments struct {
	Key         []byte
	IV          []byte
	SegmentUrls []string
}
