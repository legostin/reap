package config

import "strconv"

var defaultPortColors = map[int]string{
	// Frontend
	3000: "green", 3001: "green", 4200: "green", 5173: "green", 8080: "green",
	// Backend
	4000: "yellow", 5000: "yellow", 8000: "yellow", 8081: "yellow", 9000: "yellow",
	// Flask / Vite
	5001: "cyan", 5174: "cyan", 5175: "cyan",
	// Postgres
	5432: "magenta",
	// Redis
	6379: "red",
	// MySQL
	3306: "blue",
	// MongoDB
	27017: "blue",
	// HTTP/S
	80: "white", 443: "white",
}

// PortColor returns the color name for a given port number.
// User config overrides take precedence over defaults.
func (c Config) PortColor(port int) string {
	// Check user overrides first
	if color, ok := c.PortColors[portKey(port)]; ok {
		return color
	}
	if color, ok := defaultPortColors[port]; ok {
		return color
	}
	return "dim"
}

func portKey(port int) string {
	return strconv.Itoa(port)
}
