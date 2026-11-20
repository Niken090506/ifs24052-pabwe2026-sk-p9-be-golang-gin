package ai

type GenerateRequest struct {
	Topic            string `json:"topic"`
	Prompt           string `json:"prompt"`
	Category         string `json:"category"`
	IncludeWebSearch *bool  `json:"include_web_search"`
	UseSearchEngine  *bool  `json:"use_search_engine"`
	UseWebSearch     *bool  `json:"use_web_search"`
}

type Source struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

type Article struct {
	Title         string   `json:"title"`
	Category      string   `json:"category"`
	Excerpt       string   `json:"excerpt"`
	Content       string   `json:"content"`
	Sources       []Source `json:"sources"`
	SearchSources []Source `json:"search_sources"`
}
