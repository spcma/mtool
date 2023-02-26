package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/spf13/cast"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	path     string // 文件保存路径
	username string // 用户名
	pwd      string // 密码
	host     string // 数据库ip地址
	port     string // port
	database string
	tables   string //
)

// load package
// go get -u -v golang.org/x/tools/cmd/goimports
// go install golang.org/x/tools/cmd/goimports

// go run main.go -db [dbname] -sc [schema] -user [admin] -host [host] -port [port] -pwd [password]

func init() {
	flag.StringVar(&username, "user", "xxx", "# Database account")
	flag.StringVar(&pwd, "pwd", "xxx", "# Database password")
	flag.StringVar(&host, "host", "127.0.0.1", "# Database host")
	flag.StringVar(&port, "port", "3306", "# Database port")
	flag.StringVar(&database, "db", "xxx", "# Database name")
	flag.StringVar(&tables, "t", "", "# Table name formats such as - t user, rule, config")
	flag.StringVar(&path, "path", "xxx", "# Structure preservation path")
}

//var baseFields = []string{"id", "create_time", "update_time", "create_user", "update_user", "delete_user"}

func main() {
	flag.Parse()

	if len(host) == 0 || len(port) == 0 || len(username) == 0 || len(pwd) == 0 || len(database) == 0 {
		return
	}

	//initVar()

	//	convert aa,bb to 'aa','bb'
	tables = convtables(tables) // 转换

	fmt.Println("正在启动生成结构体...")

	execGenStruct()

	var err error
	fmt.Println("去除无用的引用包...")
	cmd := exec.Command("go imports", "-w", path)
	if err = cmd.Run(); err != nil {
		fmt.Println(err)
	}

	// format
	fmt.Println("格式化文件...")
	cmdFmt := exec.Command("gofmt", "-l", "-w", path)
	if err = cmdFmt.Run(); err != nil {
		fmt.Println(err)
	}
	fmt.Println("结构体生成完成")
}

func execGenStruct() {
	if !IsDir(path) {
		fmt.Println("目录不存在,创建目录...")
		if err := os.MkdirAll(path, 0766); err != nil {
			fmt.Println(err)
			panic(err)
		}
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
		username,
		pwd,
		host,
		cast.ToInt(port),
		database,
	)

	fmt.Println(dsn)

	orm, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	sql := `
		SELECT TABLE_NAME,column_name, data_type, column_comment, is_nullable 
		FROM information_schema.COLUMNS 
		WHERE table_schema = '%s'
	`

	searchTablesSql := fmt.Sprintf(sql, database)

	var fs []FieldInfo
	orm.Raw(searchTablesSql).Find(&fs)

	if len(fs) > 0 {
		genModelStruct(fs)
	} else {
		fmt.Println("查询不到数据")
	}
}

func genModelStruct(fs []FieldInfo) {
	var preTableName string
	var buffer bytes.Buffer

	writeField := func(fieldType, isNullable string) {
		switch fieldType {
		case "int":
			buffer.WriteString("int ")
		case "int4":
			buffer.WriteString("uint16 ")
		case "int8":
			buffer.WriteString("uint64 ")
		case "bigint":
			buffer.WriteString("uint64 ")
		case "char", "bpchar", "varchar", "longtext", "text", "tinytext":
			buffer.WriteString("string ")
		case "date", "datetime", "timestamp":
			buffer.WriteString("time.Time ")
		case "double", "float", "numeric":
			buffer.WriteString("float64 ")
		default:
			buffer.WriteString("string ")
		}
	}
	var tableCount uint16
	resetWriter := func(tableName, tableCnName string) {
		tableCount++
		buffer.Reset()
		buffer.WriteString("package model\n\n")
		buffer.WriteString("import ( \n")
		buffer.WriteString("\"database/sql\"")
		buffer.WriteString(") \n")

		buffer.WriteString("//" + tableCnName + "\n")
		buffer.WriteString("type " + fmtFieldDefine(tableName) + " struct {\n")
		preTableName = tableName
		//buffer.WriteString("Base")
		//buffer.WriteString("`map:\"dive\"`")
		//buffer.WriteString(" \n")

		fmt.Println("正在生成表结构：", tableName)
	}

	//
	writeTableNameFunc := func(tableName string) {
		structName := fmtFieldDefine(tableName)
		buffer.WriteString(fmt.Sprintf("func (*%s) TableName() string {\n", structName))
		buffer.WriteString(fmt.Sprintf("return \"%s\"\n", tableName))
		buffer.WriteString("}\n\n")

		buffer.WriteString(fmt.Sprintf("func New%s() *%s {\n", structName, structName))
		buffer.WriteString(fmt.Sprintf("return &%s{} \n", structName))
		buffer.WriteString("}")
	}

	writeFile := func() {
		buffer.WriteString("}\n")

		writeTableNameFunc(preTableName)
		filename := path + "\\" + preTableName + ".go"
		f, _ := os.Create(filename)
		_, err := f.Write([]byte(buffer.String()))
		if err != nil {
			fmt.Println(err)
			return
		}
		defer func(f *os.File) {
			err = f.Close()
			if err != nil {
				fmt.Println(err)
			}
		}(f)
	}

	//	字段个数
	fsLen := len(fs)

	for i, v := range fs {
		if v.TableName == "test" {
			fmt.Println("1")
		}
		//	如果表名为空，则重置文件头部
		if len(preTableName) == 0 {
			resetWriter(v.TableName, v.TableName)
		} else if preTableName != v.TableName || fsLen == i+1 {
			// 如果新的表名不等于旧的表名，并且为最后一个字段
			writeFile()

			if i+1 < fsLen || preTableName != v.TableName && fsLen == i+1 {
				resetWriter(v.TableName, v.TableName)
			}
		}
		// base field continue
		//if StringsContains(baseFields, strings.ToLower(v.ColumnName)) != -1 {
		//	continue
		//}

		var comment string
		if len(v.FieldCnName) > 0 {
			comment = "// " + v.FieldCnName
		}
		buffer.WriteString("" + fmtFieldDefine(v.ColumnName) + " ")
		writeField(v.FieldType, v.IsNullable)
		buffer.WriteString(fmt.Sprintf("`gorm:\"column:%s\" db:\"%s\" json:\"%s\" map:\"%s,omitempty\" form:\"%s\" label:\"%s\"` %s \n", v.ColumnName, v.ColumnName, v.ColumnName, v.ColumnName, v.ColumnName, v.FieldCnName, comment))
		if i+1 == fsLen {
			writeFile()
		}
	}
	fmt.Printf("table count: %d \n", tableCount)
}

type FieldInfo struct {
	TableName   string `gorm:"column:TABLE_NAME"`
	ColumnName  string `gorm:"column:COLUMN_NAME"`
	FieldCnName string `gorm:"column:COLUMN_COMMENT"`
	FieldType   string `gorm:"column:DATA_TYPE"`
	IsNullable  string `gorm:"column:IS_NULLABLE"`
}

func fmtFieldDefine(src string) string {
	temp := strings.Split(src, "_") // 有下划线的，需要拆分
	var str string
	for i := 0; i < len(temp); i++ {
		b := []rune(temp[i])
		for j := 0; j < len(b); j++ {
			if j == 0 {
				// 首字母大写转换
				b[j] -= 32
				str += string(b[j])
			} else {
				str += string(b[j])
			}
		}
	}
	return str
}

// json tag，首字母小写
func fmtJson(src string) string {
	temp := strings.Split(src, "_") // 有下划线的，需要拆分
	var str string
	for i := 0; i < len(temp); i++ {
		b := []rune(temp[i])
		for j := 0; j < len(b); j++ {
			if j == 0 {
				if i > 0 {
					// 首字母大写转换
					b[j] -= 32
				}
				str += string(b[j])
			} else {
				str += string(b[j])
			}
		}
	}
	return str
}

func StringsContains(array []string, val string) (index int) {
	index = -1
	for i := 0; i < len(array); i++ {
		if array[i] == val {
			index = i
			return
		}
	}
	return
}

// 用户输入转换
func convtables(tab string) string {
	if tab == "" {
		return tab
	} else {
		str_arr := strings.Split(tab, ",")
		var tabs string
		for _, v := range str_arr {
			if v != "" {
				item := fmt.Sprintf("'%s',", v)
				tabs += item
			}
		}
		// 'mall_account','mall_goods'
		return tabs[0 : len(tabs)-1]
	}
}

// IsDir 判断目录是否存在
func IsDir(fileAddr string) bool {
	s, err := os.Stat(fileAddr)
	if err != nil {
		log.Println(err)
		return false
	}
	return s.IsDir()
}
