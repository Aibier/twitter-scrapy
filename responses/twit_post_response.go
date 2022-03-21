package responses

type TwitPostResponse struct {
	Status  int                    `json:"status"`
	Message string                 `json:"message"`
	Data    map[string]interface{} `json:"data"`
}

type RecentSearchAPIResponse struct {
	Data []interface{} `json:"data"`
	Meta Meta `json:"meta"`
}

type Meta struct {
	NewestId string `json:"newest_id"`
	OldestId string `json:"oldest_id"`
	ResultCount int `json:"result_count"`
	NextToken string `json:"next_token"`
}