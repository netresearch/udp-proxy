package main

import (
	"net"
	"os"
	"strings"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
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
	log.Info().Msg("Forward targets:")
	for k, v := range config.Forwards {
		log.Info().Msgf("  %s -> %s", k, v)
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
	Port     int               `yaml:"port"`
	Forwards map[string]string `yaml:"forwards"`
}

func parseOptions() (*Config, error) {
	pflag.Int("port", 5000, "port to listen on")
	pflag.String("forwards", "", "comma separated list of forwards")

	pflag.Parse()

	viper.SetDefault("port", 5000)
	viper.SetDefault("forwards", "")

	viper.SetConfigName("udp-proxy")
	viper.SetConfigType("yml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/udp-proxy")
	viper.AddConfigPath("$HOME/.config/udp-proxy")
	viper.AddConfigPath("$HOME/.udp-proxy")

	viper.SetEnvPrefix("proxy")
	viper.BindEnv("port")
	viper.BindEnv("forwards")

	viper.BindPFlag("port", pflag.Lookup("port"))
	viper.BindPFlag("forwards", pflag.Lookup("forwards"))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}

		log.Warn().Msg("No config file found")
	}

	port := viper.GetInt("port")
	forwards := viper.GetString("forwards")

	return &Config{
		Port:     port,
		Forwards: parseForwards(forwards),
	}, nil
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

func parseForwards(raw string) map[string]string {
	forwards := make(map[string]string)

	if raw == "" {
		return forwards
	}

	rawForwards := strings.Split(raw, ",")
	for _, rawForward := range rawForwards {
		parts := strings.SplitN(rawForward, ":", 2)

		remote := parts[0]
		target := parts[1]

		forwards[remote] = target
	}

	return forwards
}
