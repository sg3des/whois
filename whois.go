package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/cheggaaa/pb"
	"github.com/jeffail/gabs"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

//"http://rest.db.ripe.net/search.json?query-string=%s&flags=resource"
var (
	ListIp     string
	All        bool
	Resultfile string

	Conf        tConf
	Presultfile *os.File
)

type tConf struct {
	Sep       string         `json:"sep"`
	Fields    []tFields      `json:"fields"`
	Length    int            `json:"length"`
	SaveOrder map[string]int `json:"saveorder"`
}

type tFields struct {
	Url    string
	Filter map[string]tFilter
}

type tFilter struct {
	Path    []string `json:"path"`
	Key     string   `json:"key"`
	Value   string   `json:"value"`
	Ret     string   `json:"ret"`
	Match   string   `json:"match,omitempty"`
	Replace string   `json:"replace,omitempty"`
	Split   string   `json:"split,omitempty"`
}

func init() {
	flag.StringVar(&ListIp, "i", "listip.csv", "set path to csv file with ip addresses")
	flag.BoolVar(&All, "a", false, "enable mode all for update information for existing element")
	flag.StringVar(&Resultfile, "o", "result.csv", "set path to file where save result")
	flag.Parse()

	b, err := ioutil.ReadFile("conf3.json")
	checkerr(err, "error read conf.json")
	err = json.Unmarshal(b, &Conf)
	checkerr(err, "error parse json structure in conf.json")

	if _, err := os.Stat(Resultfile); os.IsNotExist(err) {
		Presultfile, err = os.Create(Resultfile)
		if err != nil {
			fmt.Println("error create file result: " + Resultfile)
		}
	}
}

func main() {
	// fmt.Println("start")
	data := parseCsv(ListIp, Conf.Sep)
	bar := pb.StartNew(len(data))
	bar.ShowPercent = false
	for i := 0; i < len(data); i++ {
		bar.Increment()
		bar.Postfix(" " + data[i][0])

		if (len(data[i]) > 2 && len(data[i][1]) == 0) || (len(data[i]) == 1) || All {

			var line = make([]string, Conf.Length)
			ip := data[i][0]
			// fmt.Println(ip)
			line[Conf.SaveOrder["ip"]] = ip

			for k := 1; k < Conf.Length; k++ {
				line[k] = ""
			}
			for _, field := range Conf.Fields {
				jsonbody, err := request(ip, field.Url)
				if err != nil {
					fmt.Println(err.Error())
					continue
				}

				for key, filter := range field.Filter {
					// fmt.Println(filter)
					_, ret := search(jsonbody, filter, 0)
					// data[i][]
					// fmt.Println(ret)
					if len(filter.Split) > 0 {
						aret := regexp.MustCompile(filter.Split).Split(ret, -1)
						for j := 0; j < len(aret); j++ {
							line[Conf.SaveOrder[key+"_"+strconv.Itoa(j)]] = line[Conf.SaveOrder[key+"_"+strconv.Itoa(j)]] + strings.TrimSpace(aret[j])
						}
					} else {
						line[Conf.SaveOrder[key]] = line[Conf.SaveOrder[key]] + strings.TrimSpace(ret)
					}
				}
			}
			// fmt.Println(line)
			linesave(line, ip)
			// fmt.Println(ip)
			// os.Exit(0)
		}
	}
	bar.FinishPrint("End")
}

func parseCsv(filename string, sep string) [][]string {
	var data [][]string

	b, err := ioutil.ReadFile(filename)
	checkerr(err)

	lines := regexp.MustCompile("(\r\n)|(\r)|(\n)").Split(string(b), -1)
	for i := 0; i < len(lines); i++ {
		if len(lines[i]) > 0 {
			data = append(data, strings.Split(lines[i], sep))
		}
	}

	return data
}

func request(ip string, url string) (*gabs.Container, error) {
	var data *gabs.Container
	url = fmt.Sprintf(url, ip)

	resp, err := http.Get(url)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return data, err
	}

	jsonbody, err := gabs.ParseJSON(body)
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println("gets data is not JSON structure")
		fmt.Println(string(body))
		os.Exit(0)
		// return data, err
	}
	return jsonbody, nil
}

func search(data *gabs.Container, filter tFilter, i int) (*gabs.Container, string) {
	var ret string //returned value
	filter.Path = filter.Path[i:]
	//fmt.Println("I:::", i, filter.Path, reflect.TypeOf(data.Data()).String())
	//fmt.Println(data)
	// fmt.Println(data.Search(filter.Key).String())
	if len(filter.Value) > 0 {
		if data.Search(filter.Key).Data() == filter.Value {
			ret = data.Search(filter.Ret).Data().(string)

			if len(filter.Match) > 0 {

				if regexp.MustCompile(filter.Match).MatchString(ret) {
					if len(filter.Replace) > 0 {
						ret = regexp.MustCompile(filter.Replace).ReplaceAllString(ret, "")
					}
					return data, ret
				} else {
					ret = ""
				}

			} else {
				if len(filter.Replace) > 0 {
					ret = regexp.MustCompile(filter.Replace).ReplaceAllString(ret, "")
				}
				return data, ret
			}
		}
	}

	for j, str := range filter.Path {

		if data.Data() != nil {
			if reflect.TypeOf(data.Data()).String() == "[]interface {}" {
				children, err := data.Children()
				if err != nil {
					fmt.Println(err.Error())
				}
				for _, child := range children {
					data, ret = search(child, filter, j)
					if len(ret) > 0 {
						return data, ret
					}
				}
			} else {

				data = data.Search(str)
			}
		}
		// fmt.Println(str, data)
	}
	return data, ret
}

func linesave(line []string, ip string) {
	filedata, err := ioutil.ReadFile(Resultfile)
	checkerr(err, "error read result file")
	sline := strings.Join(line, Conf.Sep) + "\r\n"

	// fmt.Println(string(filedata), sline)

	if regexp.MustCompile("(?m)^" + regexp.QuoteMeta(ip) + "((;)|((\r\n)|\r|\n)|$)").Match(filedata) {
		// fmt.Println("match", ip)
		exp, err := regexp.Compile("(?m)^" + regexp.QuoteMeta(ip) + "((;;*.*((\r\n)|(\r)|(\n)))|((\r\n)|(\r)|(\n)))")
		checkerr(err, "error create regexp for: "+ip)

		filedata = exp.ReplaceAll(filedata, []byte(sline))
	} else {
		// fmt.Println("notmatch", ip)
		filedata = []byte(string(filedata) + sline)
	}
	// fmt.Println(string(filedata))
	// fmt.Println("========================")
	err = ioutil.WriteFile(Resultfile, filedata, 0755)
	checkerr(err, "error write to file result")
}

func checkerr(err error, str ...string) {
	if err != nil {
		if len(str) > 0 {
			fmt.Println(str)
		}
		fmt.Println(err.Error())
		os.Exit(1)
	}
}
