package entity

type IPInfo struct {
	Id        int     `json:"id"`
	IP        string  `json:"ip"`
	Continent string  `json:"continent_name"`
	Country   string  `json:"country_name"`
	Region    string  `json:"region_name"`
	City      string  `json:"city"`
	Zip       string  `json:"zip"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type IpIdDto struct {
	Id int `json:"id"`
}

type IPInfoDto struct {
	IP        string  `json:"ip"`
	Continent string  `json:"continent_name"`
	Country   string  `json:"country_name"`
	Region    string  `json:"region_name"`
	City      string  `json:"city"`
	Zip       string  `json:"zip"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}
