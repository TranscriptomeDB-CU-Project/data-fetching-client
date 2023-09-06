package utils

import (
	"arrayexpress-fetch/dtos"
	"reflect"
	"strings"

	"github.com/mitchellh/mapstructure"
)

func ExtractSDRFFileName(accessionMetadata *dtos.AccessionMetadata) []string {
	var result []string

	for _, section := range accessionMetadata.Sections.Subsection {
		if reflect.TypeOf(section).Kind() != reflect.Map {
			continue
		}

		var _section *dtos.SubsectionMetadata

		err := mapstructure.Decode(section, &_section)

		if err == nil {
			if _section.Type == "Assays and Data" {
				for _, data_file := range _section.Subsections {
					if data_file.Type == "MAGE-TAB Files" {
						for _, group_file := range data_file.Files {
							for _, file := range group_file {
								if strings.HasSuffix(file.Path, ".sdrf.txt") {
									result = append(result, file.Path)
								}
							}
						}
					}
				}
			}
		}
	}

	return result
}
