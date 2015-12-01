package transform

import (
	"encoding/json"
	"errors"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"regexp"
	"strconv"
)

func Unhtmlize(text []byte) []byte {
	re := regexp.MustCompile("( *}$)")
	text = []byte(re.ReplaceAllString(string(text), "}"))
	re = regexp.MustCompile("(\\\\u003c)")
	text = []byte(re.ReplaceAllString(string(text), "<"))
	re = regexp.MustCompile("(\\\\u003e)")
	text = []byte(re.ReplaceAllString(string(text), ">"))
	re = regexp.MustCompile("(\\\\u0026)")
	text = []byte(re.ReplaceAllString(string(text), "&"))
	return text
}

func Json2Yaml(raw []byte) ([]byte, error) {
	var empty []byte
	var data interface{}
	var err error = json.Unmarshal(raw, &data)
	if err != nil {
		return empty, err
	}
	data, err = TransformData(data)
	if err != nil {
		return empty, err
	}
	raw, err = yaml.Marshal(data)
	if err != nil {
		return empty, err
	}
	return raw, nil
}

func Yaml2JsonIndent(raw []byte) ([]byte, error) {
	var empty []byte
	var data interface{}
	var err error = yaml.Unmarshal(raw, &data)
	if err != nil {
		return empty, err
	}
	data, err = TransformData(data)
	if err != nil {
		return empty, err
	}

	raw, err = json.MarshalIndent(data, "", "  ")
	if err != nil {
		return empty, err
	}
	return raw, nil
}

func Yaml2Json(raw []byte) ([]byte, error) {
	var empty []byte
	var data interface{}
	var err error = yaml.Unmarshal(raw, &data)
	if err != nil {
		return empty, err
	}
	data, err = TransformData(data)
	if err != nil {
		return empty, err
	}

	raw, err = json.Marshal(data)
	if err != nil {
		return empty, err
	}
	return raw, nil
}

func TransformData(in interface{}) (out interface{}, err error) {
	switch in.(type) {
	case map[interface{}]interface{}:
		o := make(map[string]interface{})
		for k, v := range in.(map[interface{}]interface{}) {
			sk := ""
			switch k.(type) {
			case string:
				sk = k.(string)
			case int:
				sk = strconv.Itoa(k.(int))
			default:
				return nil, errors.New(
					fmt.Sprintf("type did not match: expected map key string or int got: %T", k))
			}
			v, err = TransformData(v)
			if err != nil {
				return nil, err
			}
			o[sk] = v
		}
		return o, nil
	case []interface{}:
		in1 := in.([]interface{})
		len1 := len(in1)
		o := make([]interface{}, len1)
		for i := 0; i < len1; i++ {
			o[i], err = TransformData(in1[i])
			if err != nil {
				return nil, err
			}
		}
		return o, nil
	default:
		return in, nil
	}
	return in, nil
}
