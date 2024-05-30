package main

import (
	"net"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	config, err := parseOptions()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to parse options")
	}
	zerolog.SetGlobalLevel(config.LogLevel)

	listener, err := net.ListenUDP("udp", &net.UDPAddr{Port: config.Port})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to listen")
	}

	log.Info().Int("port", config.Port).Msg("Listening on UDP port")
	log.Info().Msg("Forward targets:")
	for k, v := range config.Forwards {
		log.Info().Msgf("  %s -> %s", net.IP(k), v)
	}

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
	Port     int
	Forwards map[string]*net.UDPAddr
	LogLevel zerolog.Level
}

func parseOptions() (*Config, error) {
	port := pflag.Int("port", 5000, "Port to listen on")
	forwards := pflag.StringArray("forward", []string{}, "Forwards (can be specified multiple times, format: sourceIP:targetIP:targetPort)")
	rawLogLevel := pflag.String("log-level", "info", "Log level")

	pflag.Parse()

	logLevel, err := zerolog.ParseLevel(*rawLogLevel)
	if err != nil {
		return nil, err
	}

	return &Config{
		Port:     *port,
		Forwards: parseForwards(*forwards),
		LogLevel: logLevel,
	}, nil
}

func forward(config *Config, addr *net.UDPAddr, buf []byte) {
	remoteIP := addr.IP.String()

	l := log.With().Str("remote", remoteIP).Logger()

	target, ok := config.Forwards[string(addr.IP.To16())]
	if !ok {
		l.Debug().Msg("No forward for remote IP")
		return
	}
	targetAddr := target.String()

	l = l.With().Str("target", targetAddr).Logger()

	conn, err := net.DialUDP("udp", nil, target)
	if err != nil {
		l.Error().Err(err).Msg("Failed to connect to target")
		return
	}
	defer conn.Close()

	if _, err = conn.Write(buf); err != nil {
		l.Error().Err(err).Msg("Failed to write to target")
		return
	}

	l.Trace().Int("bytes", len(buf)).Msg("Forwarded")
}

func parseForwards(raw []string) map[string]*net.UDPAddr {
	forwards := make(map[string]*net.UDPAddr)

	for _, rawForward := range raw {
		l := log.With().Str("forward", rawForward).Logger()

		parts := strings.SplitN(rawForward, ":", 2)
		if len(parts) != 2 {
			l.Error().Msg("Failed to parse forward, expected 2 parts separated by :")
			continue
		}

		rawSourceIP := parts[0]
		sourceIP := net.ParseIP(rawSourceIP)
		if sourceIP == nil {
			l.Error().Msg("Failed to parse source IP")
			continue
		}

		rawTarget := parts[1]
		target, err := net.ResolveUDPAddr("udp", rawTarget)
		if err != nil {
			l.Error().Err(err).Msg("Failed to resolve target")
			continue
		}

		forwards[string(sourceIP.To16())] = target
	}

	return forwards
}
