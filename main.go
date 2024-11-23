package main

import (
	"fmt"
	"log"
	"bufio"
	"os"
	"io"
	"path"
	"strings"

	"gopkg.in/yaml.v3"
	"github.com/kolo/xmlrpc"
)

type Config struct{  //配置文件结构
	Url string `yaml:"Url"`
	UserName string `yaml:"UserName"`
	Password string `yaml:"Password"`
}

func Checkerr(err error,step string){ //检测是否出错
	if err != nil&&err != io.EOF{
		log.Fatalf("Error in %s:%v",step,err)
	}
}

func GetConfig(config *Config){  //读取配置文件
	file, err :=os.Open("config.yaml") //打开文件
	Checkerr(err,"open open the config")
	defer file.Close()

	data, err :=io.ReadAll(file) //读取文件
	Checkerr(err,"reading file")

	err = yaml.Unmarshal(data,&config) //解析yaml
	Checkerr(err,"parsing file")
}

func Getpost(post *strings.Builder,postTitle *string){
	fmt.Print("Please input the address of the markdown file:") //获取文件地址
	var filePath string
	fmt.Scanln(&filePath)
	filePath=strings.TrimRight(filePath, "'")
	filePath=strings.TrimLeft(filePath, "'")
	*postTitle = strings.TrimRight(path.Base(filePath), path.Ext(path.Base(filePath))) 

	file, err :=os.Open(filePath) //打开文件
	Checkerr(err,"opening the markdown file")
	defer file.Close()

	reader :=bufio.NewReader(file) //读取文件
	for{
		line,err := reader.ReadString('\n')
		Checkerr(err,"reading markdown file")
		line = strings.TrimRight(line,"\n")
		post.WriteString(line+"\n")
		if err == io.EOF{
			break
		}
	}
}

func GetCategory (config Config) []string{
	client,err := xmlrpc.NewClient(config.Url, nil) //创建client
	Checkerr(err,"creating client")

	args := []interface{}{1,config.UserName, config.Password} //获取内容
	var receive []map[string]interface{}
	err = client.Call("wp.getCategories", args,&receive)
	Checkerr(err,"receiving")

	categories := make([]string, 0) //处理内容
	for _, item := range receive {
		category := fmt.Sprintf("%v", item["categoryName"])
		categories = append(categories, category)
	}

	for i, category := range categories { //输出
		fmt.Printf("%v:\tName:%v\n", i, category)
	}
	return categories
}

func SendPost(config Config,postData map[string]interface{}) int {
	client, err := xmlrpc.NewClient(config.Url, nil)//创建client
	Checkerr(err,"creating client")
	
	args := []interface{}{1,config.UserName, config.Password,postData,true} //发送文章
	var postID int
	err = client.Call("metaWeblog.newPost",args,&postID)
	Checkerr(err,"Sending")

	return postID//返回postID
}

func main(){
	fmt.Println("Hello,welcome to use YPush!") 

	var config Config
	GetConfig(&config) //读取配置文件

	var post strings.Builder//获取文章内容
	var postTitle string
	Getpost(&post,&postTitle) 

	category := make([]string,0) //获取并选择选择分类
	var num int
	fmt.Println("There are all the category in your blog:")
	categories :=GetCategory(config)
	fmt.Print("Please input  IDs that you need(separated by spaces,end with -1):")
	for fmt.Scanf("%d",&num);num!=-1;fmt.Scanf("%d",&num){
		category = append(category, categories[num])
	}

	var str,keywords string//选择标签
	fmt.Print("Please input the tags (separated by spaces,end with -1): ")
	for fmt.Scanf("%s",&str);str!="-1";fmt.Scanf("%s",&str){
		keywords = keywords +str+","
	}
	keywords=strings.TrimRight(keywords, ",")
	
	line := strings.Repeat("-", 50)//确认,发送
	fmt.Println(line)
	fmt.Println("The artical title:",postTitle)
	fmt.Print("The category:")
	for _,value:=range category{
		fmt.Print(value," ")
	}
	fmt.Println("\nThe  keywords:",keywords)
	fmt.Println(line)
	fmt.Print("Enter \"Y\" to send the post:")
	var ch int
	fmt.Scanf("%c",&ch)
	if ch != 'Y'&&ch != 'y' {
		os.Exit(1)
	}

	postData := map[string]interface{}{ //发送
		"title":       postTitle,
		"description": post.String(),
		"categories":  category, 
		"mt_keywords":keywords,
	}
	postID:=SendPost(config,postData)

	fmt.Printf("Article posted successfully,ID: %d\n", postID)
}