package libs

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"strings"
)

type handbrakePresetFile struct {
	PresetList []struct {
		PresetName string `json:"PresetName"`
	} `json:"PresetList"`
}

// ProfilesByCategory maps category name -> list of preset names.
type ProfilesByCategory map[string][]string

func GetHandBrakeProfiles() (ProfilesByCategory, error) {
	result := make(ProfilesByCategory)

	if err := loadBuiltinProfiles(result); err != nil {
		return nil, err
	}

	loadCustomProfiles("config/custom-presets.json", result)

	return result, nil
}

func loadBuiltinProfiles(result ProfilesByCategory) error {
	cmd := exec.Command("HandBrakeCLI", "--preset-list")
	out, err := cmd.CombinedOutput()
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			return err
		}
		// non-zero exit is normal for some HandBrakeCLI versions; continue with output
	}

	var currentCategory string
	scanner := bufio.NewScanner(bytes.NewReader(out))
	for scanner.Scan() {
		line := scanner.Text()
		// Category lines: no indent, end with "/"
		if !strings.HasPrefix(line, " ") && strings.HasSuffix(line, "/") {
			currentCategory = strings.TrimSuffix(line, "/")
			continue
		}
		// Preset lines: exactly 4 spaces indent (description lines have 8+)
		if strings.HasPrefix(line, "    ") && !strings.HasPrefix(line, "        ") && currentCategory != "" {
			name := strings.TrimSpace(line)
			if name != "" {
				result[currentCategory] = append(result[currentCategory], name)
			}
		}
	}
	return scanner.Err()
}

func loadCustomProfiles(path string, result ProfilesByCategory) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}

	var presetFile handbrakePresetFile
	if err := json.Unmarshal(data, &presetFile); err != nil {
		return
	}

	for _, p := range presetFile.PresetList {
		if p.PresetName != "" {
			result["Custom"] = append(result["Custom"], p.PresetName)
		}
	}
}

func transcode(config Config, path string, outputPath string) error {

	args := []string{"-i", path, "-o", outputPath}

	if config.HandbrakeCategory == "Custom" {
		args = append(args, "--preset-import-file", "config/custom-presets.json", "-Z", config.HandbrakeProfile)
	} else {
		args = append(args, "-Z", config.HandbrakeProfile)
	}

	cmd := exec.Command("HandBrakeCLI", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
