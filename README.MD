#WHOIS DATA FOR IP
	This script is written on GO(golang)
	
Data gathering from any online json service by file instruction and save this into CSV file

#Example instruction

File instruction allow getting any data from any json structure.

instruction this is JSONstruct file:

	{
		"Sep": ";",
		"Fields": [{
			"Url": "http://rest.db.ripe.net/search.json?query-string=%s",
			"Filter": {
				"inet": {
					"path": ["objects", "object", "attributes", "attribute", "name"],
					"key": "name",
					"value": "inetnum",
					"ret": "value",
					"split": "-"
				},
				"descr": {
					"path": ["objects", "object", "attributes", "attribute", "name"],
					"key": "name",
					"value": "descr",
					"ret": "value"
				},
				"domain": {
					"path": ["objects", "object", "attributes", "attribute", "name"],
					"key": "name",
					"value": "remarks",
					"ret": "value",
					"match": "@",
					"replace": ".*@"
				}
			}
		}],
		"Length": 9,
		"SaveOrder": {
			"ip": 0,
			"inet_0": 1,
			"inet_1": 2,
			"domain": 3
		}
	}
	
path - name field from json struct

key:value - if find field with name = "key" and field with value = "value" - return "ret" field value

	["name":"inetnum","value":"10.0.0.0 - 10.255.255.255"]
	
split - split recieved string and save part by name_i

match - sometimes under filter fit many fields - by match their may sift inappropriate

replace - regexp expression for replace recieved string

	
#FLAGS

(-l)input file - CSV or ip list separate line ending((\r\n)|(\r)|(\n))

(-r)output file - CSV

you may set output file same input - the result will be stored properly appending fields

(-a)you may set ALL mode - this mode rewrite existing line - default false
