
  
    
# Config      
 The objective of the config package is to allow a client to tag a struct with (optional)default values and environment variables and then the config package will use viper to populate a struct.     
    
## Tag descriptions
 |  Tag|Mandatory  |Description|    
|--|--|--|    
| env |Y  |The environment variable the field value will be retrieved from if the environment variable is present. N.B. It takes priority over any other tag.|    
| file |N  |The path to the file to take the default value from. This can be useful for things like passwords where they can be taken from docker / k8s secrets. N.B. If this tag is present and the file can be successfully processed then the default is redundant - if the value can not be read from the file and the default is present then the default is used.|  
| default |N  |The default value which will be used if no environment variable is present. N.B. If this is not populated then the default for the type will be used i.e. 0 for int, "" for string, false for bool etc etc.|    
    
N.B. A nested struct does not need tags associated with it. Tags are only required for non struct entries.    
N.N.B. Nested structs will use the "squash" property by default. That means no mapstructure tag is required.    
    
## Struct API 
``` LoadViperConfig(cfg interface{}) error ``` 
### Parameters
 |  Name|Input/Output  |Type|Description|    
|--|--|--|--|    
| cfg |input/output  |interface{}  | A pointer to the config structure which will be populated on return.  |    
| - |output  |error.   |This will hold an error if there are any issues with the tagging or viper was unable to unmarshal |    

## File API
``` LoadViperConfigFromFile(filename string, cfg interface{}) error ``` 
   ### Parameters
 |  Name|Input/Output  |Type|Description|    
|--|--|--|--|    
| filename |input  |string  | The filename of the config file - this could be yaml, json, xml, env or ini. (i.e. anything which Viper supports)  |
| cfg |input/output  |interface{}  | A pointer to the config structure which will be populated on return. |    

## Reader API
``` LoadViperConfigFromReader(reader io.Reader, cfg interface{}, cfgType string) error ``` 
   ### Parameters
 |  Name|Input/Output  |Type|Description|    
|--|--|--|--|    
| reader |input  |io.Reader  | An io.reader which can be attached to a file, a string, a byte buffer etc etc  |
| cfg |input/output  |interface{}  | A pointer to the config structure which will be populated on return.  |   
 | cfgType |input  |string  | The type of config contained within the reader - N.B. This is the range of supported viper types and can be json, yaml/yml, hcl, props, env, toml, ini. See viper docs for the full set.  |   
| - |output  |error.   |This will hold an error if there are any issues with the tagging or viper was unable to unmarshal |     
        
### Examples 
#### Simple Struct 
``` 
type TestStruct struct {       
   TestVal string `env:"TEST_VAL" default:"abc"`   
   TestFileVal string `env:"TEST_FILE_VAL" file:"/tmp/test_file_val"`  
} 
``` 
#### Nested Structs
 ```  
//InnerStruct is the nested struct 
type InnerStruct struct {    
 TestNestInner string `env:"TEST_NEST_INNER" default:"inner"`  
}
    
//OuterStruct is the struct containing the nested struct 
type OuterStruct struct {    
 TestNestOuter string `env:"TEST_NEST_OUTER" default:"outer"` InnerStruct  
}
```
N.B. This library only supports anonymous nested structs.
#### Handling the response 
The struct passed in will be populated upon return if no error.
``` 
var testStruct TestStruct
err := config.LoadViperConfig(&testStruct) 
if err != nil {      
	//Error handling code 
} 
```