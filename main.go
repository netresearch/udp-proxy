package main

import (
	"flag"
	"net"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v3"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}).Level(zerolog.TraceLevel)
	config, err := parseOptions()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse options")
	}

	listener, err := net.ListenUDP("udp", &net.UDPAddr{Port: config.Port})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen")
	}

	log.Info().Int("port", config.Port).Msg("Listening on UDP port")

	for {
		buf := make([]byte, 1024)
		n, addr, err := listener.ReadFromUDP(buf)
		if err != nil {
			log.Error().Err(err).Msg("Failed to read from UDP")
			continue
		}

		forward(config, addr, buf[:n])
	}
}

type Config struct {
	Port     int               `yaml:"port"`
	Forwards map[string]string `yaml:"forwards"`
}

func parseOptions() (*Config, error) {
	configPath := flag.String("config", "config.yml", "Path to the config file")

	flag.Parse()

	fd, err := os.Open(*configPath)
	if err != nil {
		return nil, err
	}
	defer fd.Close()

	var config Config
	if err = yaml.NewDecoder(fd).Decode(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func forward(config *Config, addr *net.UDPAddr, buf []byte) {
	remoteIP := addr.IP.String()

	l := log.With().Str("remote", remoteIP).Logger()

	target, ok := config.Forwards[remoteIP]
	if !ok {
		l.Warn().Msg("No forward for remote IP")
		return
	}

	l = l.With().Str("target", target).Logger()

	targetAddr, err := net.ResolveUDPAddr("udp", target)
	if err != nil {
		l.Error().Err(err).Msg("Failed to resolve target")
		return
	}

	conn, err := net.DialUDP("udp", nil, targetAddr)
	if err != nil {
		l.Error().Err(err).Msg("Failed to connect to target")
		return
	}
	defer conn.Close()

	if _, err = conn.Write(buf); err != nil {
		l.Error().Err(err).Msg("Failed to write to target")
		return
	}

	l.Info().Int("bytes", len(buf)).Msg("Forwarded")
}
