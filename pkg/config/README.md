
# Config  
  
The objective of the config package is to allow a client to tag a struct with (optional)default values and environment variables and then the config package will use viper to populate a struct. 

## Tag descriptions
|  Tag|Mandatory  |Description|
|--|--|--|
| env |Y  |The environment variable the field value will be retrieved from if the environment variable is present.|
| default |N  |The default value which will be used if no environment variable is present. N.B. If this is not populated then the default for the type will be used i.e. 0 for int, "" for string, false for bool etc etc.|

N.B. A nested struct does not need tags associated with it. Tags are only required for non struct entries.
N.N.B. Nested structs will use the "squash" property by default. That means no mapstructure tag is required.

## API
```
LoadViperConfig(cfg interface{}) (interface{}, error)
```
### Parameters
|  Name|Input/Output  |Type|Description|
|--|--|--|--|
| cfg |input  |interface{}  | All that is a required is an empty struct of the type you want the config read into.  |
| - |output  |interface{}.  | This output interface should hold the populated configuration struct.(or structs if nested)  |
| - |output  |error.   |This will hold an error if there are any issues with the tagging or viper was unable to unmarshal |


    
### Examples
#### Simple Struct
```
type TestStruct struct {   
		TestVal string `env:"TEST_VAL" default:"abc"` 
}  
```
#### Nested Struct
```
//InnerStruct is the nested struct
type InnerStruct struct {
	TestNestInner string `env:"TEST_NEST_INNER" default:"inner"`
}

//OuterStruct is the struct containing the nested struct
type OuterStruct struct {
	TestNestOuter string `env:"TEST_NEST_OUTER" default:"outer"`
	InnerStruct
}
```
#### Handling the response
The return from LoadViperConfig will be an interface however there is an expectation that it should be of the same type as the empty struct which was passed in. The code below is illustrative of how the return should be handled. N.B. the ok variable and check do not need to be used but if they are not used and for some reason the types do not match the code will panic. 
```
cfgInt, err := config.LoadViperConfig(TestStruct{})  
if err != nil {  
   //Error handling code
}  
cfg, ok := cfgInt.(TestStruct)  
if !ok {  
   //Further error handling  
}
```