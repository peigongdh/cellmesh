package discovery

import (
	"fmt"
	"reflect"

	"github.com/davyxu/cellnet/util"
)

// KV中的Value最大不超过512K,
const (
	// 不能直接保存二进制，底层用Json转base64，base64的二进制比原二进制要大最终二进制不到512K就会达到限制
	PackedValueSize = 300 * 1024
)

type rawGetter interface {
	// 获取原始值
	GetRawValue(key string) ([]byte, error)
	GetValueDirect(key string, valuePtr interface{}) error
}

func getMultiKey(sd rawGetter, key string) (ret []string) {

	mainKey := key

	ret = append(ret, mainKey)

	for i := 1; ; i++ {

		key = fmt.Sprintf("%s.%d", mainKey, i)

		_, err := sd.GetRawValue(key)
		if err != nil && err.Error() == "value not exists" {
			return
		}

		ret = append(ret, key)
	}

}

// compress value按 key, key.1, key.2 ... 保存
func SafeSetValue(sd Discovery, key string, value interface{}, compress bool) error {
	if compress {
		cData, err := util.CompressBytes(value.([]byte))
		if err != nil {
			return err
		}

		if len(cData) >= PackedValueSize {

			for _, multiKey := range getMultiKey(sd.(rawGetter), key) {

				err := sd.DeleteValue(multiKey)
				if err != nil {
					fmt.Printf("delete kv error, %s\n", err)
				}
			}

			var pos = PackedValueSize

			err = sd.SetValue(key, cData[:pos])
			if err != nil {
				return err
			}

			index := 1
			for len(cData)-pos > PackedValueSize {

				multiKey := fmt.Sprintf("%s.%d", key, index)
				err = sd.SetValue(multiKey, cData[pos:pos+PackedValueSize])
				if err != nil {
					return err
				}
				pos += PackedValueSize
				index++
			}

			if len(cData)-pos > 0 {
				multiKey := fmt.Sprintf("%s.%d", key, index)
				err = sd.SetValue(multiKey, cData[pos:])
				if err != nil {
					return err
				}
			}

			return nil

		} else {
			return sd.SetValue(key, cData)
		}

	} else {
		return sd.SetValue(key, value)
	}
}

func SafeGetValue(sd Discovery, key string, valuePtr interface{}, decompress bool) error {

	rg := sd.(rawGetter)

	var (
		finalData []byte
		err       error
	)

	if decompress {

		var data []byte
		for _, multiKey := range getMultiKey(rg, key) {

			var partData []byte
			err := rg.GetValueDirect(multiKey, &partData)
			if err != nil {
				return err
			}

			data = append(data, partData...)
		}

		finalData, err = util.DecompressBytes(data)

		if err != nil {
			return err
		}

		reflect.ValueOf(valuePtr).Elem().Set(reflect.ValueOf(finalData))
	} else {
		err = rg.GetValueDirect(key, &finalData)

		if err != nil {
			return err
		}
	}

	return nil
}
