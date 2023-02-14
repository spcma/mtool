package main

import (
	"bytes"
	"flag"
	"fmt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	path       string // 文件保存路径
	dbUser     string // 用户名
	dbPwd      string // 密码
	dbHost     string // 数据库ip地址
	dbPort     string // port
	dbName     string
	dbSchema   string //
	tableNames string //
)

// use
// go get -u -v golang.org/x/tools/cmd/goimports
// go install golang.org/x/tools/cmd/goimports

// go run db2struct.go -db [dbname] -sc [schema] -user [admin] -host [host] -port [port] -pwd [password]

// 注释：参数信息
//
// -host host 改为自己数据库的地址 （ 默认 127.0.0.1）
//
// -port port 改为自己数据库的端口 （ 默认 5432）
//
// -user user 改为自己数据库的账号 （ 默认 postgres）
//
// -pwd pwd 改为自己数据库的密码 （ 默认 postgres）
//
// -db dbname 改为自己数据库的名称 （必填）
//
// -sc schema 改为自己数据库的名称 （public）
//
// -path ./models 改为存放路径 (可选默认为./models )
//
// -t account,user 改为要生成的表名称、可多个 (可选默认全部生成)
func init() {
	flag.StringVar(&dbHost, "host", "127.0.0.1", "# Database host")
	flag.StringVar(&dbPort, "port", "3306", "# Database port")
	flag.StringVar(&dbName, "db", "test", "# Database name")
	flag.StringVar(&dbSchema, "sc", "test", "# schema name")
	flag.StringVar(&dbUser, "user", "admin", "# Database account")
	flag.StringVar(&dbPwd, "pwd", "admin@123", "# Database password")
	flag.StringVar(&tableNames, "t", "", "# Table name formats such as - t user, rule, config")
	flag.StringVar(&path, "path", "./model", "# Structure preservation path")
}

var baseFields = []string{"id", "create_time", "update_time", "create_user", "update_user", "delete_user"}

func initVar() {
	// 文件保存路径 改第二个值
	path = "./models"
	// 用户名
	dbUser = "admin"
	// 密码
	dbPwd = "admin@123"
	// 数据库ip地址
	dbHost = "127.0.0.1"
	dbPort = "3306"
	dbName = "test"
}

func main() {
	flag.Parse()
	// initVar()
	if len(dbName) == 0 {
		flag.Usage()
		return
	}

	//	convert aa,bb to 'aa','bb'
	tableNames = convtables(tableNames) // 转换

	fmt.Println("地址:", dbHost)
	fmt.Println("端口:", dbPort)
	fmt.Println("数据库:", dbName)
	fmt.Println("模式名:", dbSchema)
	fmt.Println("数据库账号:", dbUser)
	fmt.Println("数据库密码:", dbPwd)
	fmt.Println("结构体保存路径:", path)
	fmt.Println("指定数生成据表:", tableNames)
	fmt.Println("正在启动生成结构体....")

	//	core generate function
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
		"spcmaadmin",
		"admin@123",
		"101.35.184.189",
		13306,
		"spcma",
	)

	orm, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	var fs []FieldInfo
	orm.Raw("SELECT TABLE_NAME,column_name, data_type, column_comment, is_nullable FROM information_schema.COLUMNS where table_schema = 'spcma'").Find(&fs)

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
			// buffer.WriteString("time.Time ")
		case "double", "float", "numeric":
			buffer.WriteString("float64 ")
		default:
			// 其他类型当成string处理
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
		f.Write([]byte(buffer.String()))
		f.Close()
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
