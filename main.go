package main

import (
	"buildinInstallAssistant/common/crypto/myaes"
	"buildinInstallAssistant/common/http"
	"crypto/aes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/jony-lee/go-progress-bar"
	"log"
	"math"
	"os"
	"slices"
	"time"
)

var aesKey = "abcdefghig123456"
var authBaseUrl = "http://119.39.84.24:6363"

var projectBaseUrl = "http://119.39.84.24:8200"

func initLog() {
	// 创建、追加、读写，777，所有权限
	f, err := os.OpenFile("log.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		return
	}

	log.SetOutput(f)
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
}

func init() {
	initLog()
}

func main() {
	var username, password string
	fmt.Println("欢迎使用长沙市装配式建筑项目信息采集平台预制构件自动安装助手")
	fmt.Print("\r请输入登录用户名：")
	if _, err := fmt.Scanln(&username); err != nil {
		fmt.Println("接收用户名错误", err.Error())
		return
	}
	clearConsole(-1)
	fmt.Print("\r请输入登录密码：")
	if _, err := fmt.Scanln(&password); err != nil {
		fmt.Println("接收密码错误", err.Error())
		return
	}
	log.Printf("输入的用户名密码为：%s:%s", username, password)
	loginUser, err := login(username, password)
	if err != nil {
		fmt.Println("登录失败！", err.Error())
		log.Println("登录失败！", err.Error())
		return
	}
	clearConsole(-1)
	fmt.Printf("\r当前登录用户为：%s\n", loginUser.UserName)
	//打印空行
	fmt.Println("")
	projectResult, err := getProjectList(1)
	if err != nil {
		fmt.Println("获取项目列表失败！", err.Error())
		log.Println("获取项目列表失败！", err.Error())
		return
	}
	projectTotalCount := projectResult.TotalCount
	if projectTotalCount == 0 {
		fmt.Println("当前没有项目！")
	}
	projectRows := projectResult.Table.Rows
	for i, project := range projectRows {
		fmt.Println("[", i+1, "]、", project.Name)
	}
	fmt.Print("输入项目列表前的序号选择要操作的项目(默认为 1)：")
	var selectIndex int
	if _, err := fmt.Scanln(&selectIndex); err != nil {
		log.Println("输入的项目序号不支持：", err.Error())
		selectIndex = 1
	}
	for selectIndex < 1 || selectIndex > len(projectRows) {
		clearConsole(-1)
		fmt.Print("\r输入项目列表前的序号选择要操作的项目(默认为 1)：")
		if _, err := fmt.Scanln(&selectIndex); err != nil {
			log.Println("输入的项目序号不支持：", err.Error())
			selectIndex = 1
		}
	}
	project := projectRows[selectIndex-1]
	fmt.Println("数据准备中……")
	installResult, err := getInstallList(project.Id, 1, 1)
	if err != nil {
		fmt.Println("数据准备失败！", err.Error())
		log.Println("数据准备失败！", err.Error())
		return
	}
	installTotalCount := installResult.TotalCount
	bar := progress.New(int64(installTotalCount))
	failMsgSlice := make([]string, 0)
	failIdSlice := make([]string, 0)
	for i := 0; installResult.TotalCount > len(failIdSlice) && i < installTotalCount; i++ {
		pageNumber := int(math.Ceil(float64(len(failIdSlice)) / float64(50)))
		installResult, err = getInstallList(project.Id, pageNumber, 50)
		if err != nil {
			fmt.Println("获取要安装的构件列表失败！", err.Error())
			log.Println("获取要安装的构件列表失败！", err.Error())
			return
		}
		installRows := installResult.Table.Rows
		for _, row := range installRows {
			if slices.Contains(failIdSlice, row.Id) {
				continue
			}
			if _, err := installOne(row); err != nil {
				failMsg := fmt.Sprintf("构件[%s]安装失败,构件RFID编号为[%s]!", row.Productname, row.Rfidnum)
				log.Println(failMsg)
				failMsgSlice = append(failMsgSlice, failMsg)
				failIdSlice = append(failIdSlice, row.Id)
			}
			bar.Done(1)
			// 休眠50毫秒，防止请求太快把服务器搞崩
			time.Sleep(50 * time.Millisecond)
		}
	}
	bar.Finish()
	if len(failMsgSlice) == 0 {
		fmt.Println("所有构件安装完成！")
	} else {
		fmt.Println("构件自动安装完成，但有部分构件安装失败，失败的构件信息如下：")
		for i, s := range failMsgSlice {
			fmt.Println(i+1, ". ", s)
		}
	}
	fmt.Println("按ENTER键关闭本窗口……")
	_, _ = fmt.Scanln()

}

func clearConsole(lineNumber int) {
	if lineNumber == 0 {
		return
	}
	if lineNumber > 0 {
		for i := 0; i < lineNumber; i++ {
			fmt.Print("\033[B")
			blankStr := ""
			for j := 0; j < 100; j++ {
				blankStr += " "
			}
			fmt.Print("\r", blankStr)
		}
	} else {
		for i := lineNumber; i < 0; i++ {
			fmt.Print("\033[1A")
			blankStr := ""
			for j := 0; j < 100; j++ {
				blankStr += " "
			}
			fmt.Print("\r", blankStr)
		}
	}
}

type LoginData struct {
	UserCode             string `json:"userCode"`
	UserPwd              string `json:"userPwd"`
	ChinaHouseUserCode   string `json:"chinaHouseUserCode"`
	NeedRememberUserInfo string `json:"needRememberUserInfo"`
}

type Org struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Lon  string `json:"lon"`
	Id   string `json:"id"`
	Lat  string `json:"lat"`
}
type Role struct {
	Code string `json:"code"`
	Name string `json:"name"`
	Id   string `json:"id"`
}

type Response[T any] struct {
	Result  T      `json:"result"`
	Code    string `json:"code"`
	Message string `json:"message"`
}
type LoginUser struct {
	OrgList    []Org   `json:"orgList"`
	UserName   string  `json:"userName"`
	RoleList   []Role  `json:"roleList"`
	CSessionId string  `json:"cSessionId"`
	UserId     string  `json:"userId"`
	UserCode   string  `json:"userCode"`
	Token      *string `json:"token"`
}

func login(username string, password string) (*LoginUser, error) {
	if username == "" || password == "" {
		return nil, errors.New("用户名或密码不能为空")
	}
	pwd, err := encryptPwd(password)
	if err != nil {
		return nil, err
	}
	params := LoginData{UserCode: username, UserPwd: pwd, NeedRememberUserInfo: "N"}
	loingData := map[string]interface{}{"requestParam": params}
	resultStr, err := http.PostForm(authBaseUrl+"/integrateUserNcpService/login.action", loingData)
	if err != nil {
		log.Println("登录失败!", err.Error())
		return nil, err
	}
	var response []Response[LoginUser]
	err = json.Unmarshal([]byte(*resultStr), &response)
	if err != nil {
		log.Println("反序列化登录响应信息失败", err.Error())
		return nil, err
	}
	if len(response) == 0 || response[0].Code != "000" {

		return nil, errors.New(response[0].Message)
	}
	return &response[0].Result, nil
}

type ProjectResultRow struct {
	Modelname         string `json:"modelname"`
	County            string `json:"county"`
	Allarea           string `json:"allarea"`
	ProjectStateXid   string `json:"project_state_xid"`
	ComtypeRegionid   string `json:"comtype_regionid"`
	OrgXid            string `json:"org_xid"`
	Isdeleted         string `json:"isdeleted"`
	ProjectnatureXid  string `json:"projectnature_xid"`
	AllAddress        string `json:"allAddress"`
	Province          string `json:"province"`
	Licensenum        string `json:"licensenum"`
	Id                string `json:"id"`
	Lat               string `json:"lat"`
	MonitorUrl        string `json:"monitor_url"`
	Projectstyle      string `json:"projectstyle"`
	ApprovalDate      string `json:"approval_date"`
	Landtransferratio string `json:"landtransferratio"`
	ChargePerson      string `json:"charge_person"`
	Modelid           string `json:"modelid"`
	Permit1Date       string `json:"permit1_date"`
	RegionXid         string `json:"region_xid"`
	DrawingsDate      string `json:"drawings_date"`
	Projectnature     string `json:"projectnature"`
	AreaAssembly      string `json:"area_assembly"`
	CountyXid         string `json:"county_xid"`
	Jsdw              string `json:"jsdw"`
	IsEpc             string `json:"is_epc"`
	Linkman           string `json:"linkman"`
	DataUniqueCode    string `json:"data_unique_code"`
	Name              string `json:"name"`
	Planid            string `json:"planid"`
	IsFufei           string `json:"is_fufei"`
	Status            string `json:"status"`
	Note              string `json:"note"`
	Upallarea         string `json:"upallarea"`
	Permit2Date       string `json:"permit2_date"`
	Code              string `json:"code"`
	Deletetime        string `json:"deletetime"`
	City              string `json:"city"`
	Modifytime        string `json:"modifytime"`
	Modifyuser        string `json:"modifyuser"`
	Statustext        string `json:"statustext"`
	Description       string `json:"description"`
	BuilderLicense    string `json:"builder_license"`
	Lon               string `json:"lon"`
	ProjectImage      string `json:"project_image"`
	IsDeactivate      string `json:"is_deactivate"`
	ProjectstyleXid   string `json:"projectstyle_xid"`
	Allinvest         string `json:"allinvest"`
	Scope             string `json:"scope"`
	Planname          string `json:"planname"`
	PaymentProof      string `json:"payment_proof"`
	ManagerXid        string `json:"manager_xid"`
	CreateuserXid     string `json:"createuser_xid"`
	Createtime        string `json:"createtime"`
	Address           string `json:"address"`
	Manager           string `json:"manager"`
	ProjectState      string `json:"project_state"`
	ImgFile           string `json:"img_file"`
	ChargePersonXid   string `json:"charge_person_xid"`
	IsselfBuild       string `json:"isself_build"`
	ProvinceXid       string `json:"province_xid"`
	Createuser        string `json:"createuser"`
	Unitprojectnum    string `json:"unitprojectnum"`
	LonLat            string `json:"lon_lat"`
	Linktel           string `json:"linktel"`
	AcceptDate        string `json:"accept_date"`
	CityXid           string `json:"city_xid"`
	File4             string `json:"file4"`
	ModifyuserXid     string `json:"modifyuser_xid"`
	JsdwXid           string `json:"jsdw_xid"`
	File5             string `json:"file5"`
	LandtenureDate    string `json:"landtenure_date"`
	File2             string `json:"file2"`
	Progress          string `json:"progress"`
	File3             string `json:"file3"`
	BiddingProjectXid string `json:"bidding_project_xid"`
	File0             string `json:"file0"`
	File1             string `json:"file1"`
	Permit3Date       string `json:"permit3_date"`
	UseDate           string `json:"use_date"`
}
type TableResult[T any] struct {
	PageSize    int `json:"pageSize"`
	TotalCount  int `json:"totalCount"`
	CurrentPage int `json:"currentPage"`
	Table       struct {
		Rows []*T `json:"rows"`
	} `json:"table"`
}

func getProjectList(pageNum int) (*TableResult[ProjectResultRow], error) {
	if pageNum == 0 {
		pageNum = 1
	}
	params := map[string]interface{}{"filter": "", "currentPage": pageNum, "pageSize": 16, "total": 0}
	postData := map[string]interface{}{"requestParam": params}
	resultStr, err := http.PostForm(projectBaseUrl+"/pjDataService/getProjectListByOrgId.action", postData)
	if err != nil {
		log.Println("从服务器获取项目信息失败！", err.Error())
		return nil, err
	}
	var response []Response[TableResult[ProjectResultRow]]
	err = json.Unmarshal([]byte(*resultStr), &response)
	if err != nil {
		log.Println("反序列化项目列表失败", err.Error())
		return nil, err
	}
	if len(response) == 0 || response[0].Code != "000" {

		return nil, errors.New(response[0].Message)
	}
	return &response[0].Result, nil
}

type InstallRow struct {
	SysCode           string `json:"sys_code"`
	Pointx            string `json:"pointx"`
	Pointy            string `json:"pointy"`
	Pointtox          string `json:"pointtox"`
	Pointtoy          string `json:"pointtoy"`
	ComptypeOrgXid    string `json:"comptype_org_xid"`
	Project           string `json:"project"`
	Axis              string `json:"axis"`
	OrgXid            string `json:"org_xid"`
	Specifications    string `json:"specifications"`
	Isdeleted         string `json:"isdeleted"`
	Producttype       string `json:"producttype"`
	Comtype           string `json:"comtype"`
	WallArea          string `json:"wall_area"`
	Upfile            string `json:"upfile"`
	Id                string `json:"id"`
	PjProjectXid      string `json:"pj_project_xid"`
	Createusername    string `json:"createusername"`
	Lat               string `json:"lat"`
	SpecialFloor      string `json:"special_floor"`
	ConcreteDosage    string `json:"concrete_dosage"`
	ManufacturerXid   string `json:"manufacturer_xid"`
	Weight            string `json:"weight"`
	Unitproject       string `json:"unitproject"`
	IsSpecialBuilding string `json:"is_special_building"`
	Volume            string `json:"volume"`
	ConcretegradeXid  string `json:"concretegrade_xid"`
	Ischeckpass       string `json:"ischeckpass"`
	UnitprojectXid    string `json:"unitproject_xid"`
	Installtime       string `json:"installtime"`
	DirectionXid      string `json:"direction_xid"`
	InstallleaderXid  string `json:"installleader_xid"`
	Note              string `json:"note"`
	Code              string `json:"code"`
	Modifytime        string `json:"modifytime"`
	RfidnumXid        string `json:"rfidnum_xid"`
	Concretegrade     string `json:"concretegrade"`
	Installleader     string `json:"installleader"`
	Lon               string `json:"lon"`
	ProductCode       string `json:"product_code"`
	Manufacturer      string `json:"manufacturer"`
	Orgname           string `json:"orgname"`
	Productname       string `json:"productname"`
	Rfidnum           string `json:"rfidnum"`
	Floor             string `json:"floor"`
	CreateuserXid     string `json:"createuser_xid"`
	ProductXid        string `json:"product_xid"`
	Direction         string `json:"direction"`
	Createtime        string `json:"createtime"`
	Blocks            string `json:"blocks"`
	ProjectXid        string `json:"project_xid"`
	Url               string `json:"url"`
	Isbatch           string `json:"isbatch"`
	ModifyuserXid     string `json:"modifyuser_xid"`
}

type InstallQuery struct {
	UnitprojectXid string `json:"unitproject_xid"`
	Floor          string `json:"floor"`
	Filter         string `json:"filter"`
	Producttype    string `json:"producttype"`
	PersonXid      string `json:"person_xid"`
	WfIsend        string `json:"wf_isend"`
	Isbatch        string `json:"isbatch"`
	CurrentPage    int    `json:"currentPage"`
	PageSize       int    `json:"pageSize"`
	Total          int    `json:"total"`
	ProjectXid     string `json:"project_xid"`
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
}

func getInstallList(projectId string, pageNumber int, pageSize int) (*TableResult[InstallRow], error) {
	if pageSize == 0 {
		pageSize = 50
	}
	if pageNumber == 0 {
		pageNumber = 1
	}
	params := InstallQuery{ProjectXid: projectId, WfIsend: "N", Isbatch: "N", PageSize: pageSize, CurrentPage: pageNumber}
	postData := map[string]interface{}{"requestParam": params}
	resultStr, err := http.PostForm(projectBaseUrl+"/installService/getInstallList.action", postData)
	if err != nil {
		log.Println("从服务器获取构件列表失败！", err.Error())
		return nil, err
	}
	var response []Response[TableResult[InstallRow]]
	err = json.Unmarshal([]byte(*resultStr), &response)
	if err != nil {
		log.Println("反序列化构件列表失败", err.Error())
		return nil, err
	}
	if len(response) == 0 || response[0].Code != "000" {

		return nil, errors.New(response[0].Message)
	}
	return &response[0].Result, nil
}

type InstallInfo struct {
	Ids            string `json:"ids"`
	Unitproject    string `json:"unitproject"`
	UnitprojectXid string `json:"unitproject_xid"`
	Floor          string `json:"floor"`
	Pointx         string `json:"pointx"`
	Pointtox       string `json:"pointtox"`
	Pointy         string `json:"pointy"`
	Pointtoy       string `json:"pointtoy"`
	DirectionXid   string `json:"direction_xid"`
	Code           string `json:"code"`
	Note           string `json:"note"`
	Upfile         string `json:"upfile"`
	IsHasFile      string `json:"isHasFile"`
	ProjectXid     string `json:"project_xid"`
	Isbatch        string `json:"isbatch"`
	InstallTime    string `json:"install_time"`
}

func installOne(installRow *InstallRow) (bool, error) {
	params := InstallInfo{ProjectXid: installRow.PjProjectXid, Ids: installRow.Id, Isbatch: "N", InstallTime: time.Now().Format("2006-01-02")}
	postData := map[string]interface{}{"requestParam": params}
	resultStr, err := http.PostForm(projectBaseUrl+"/installService/handleInstallList.action", postData)
	if err != nil {
		log.Printf("构件[%s](%s)安装请求失败！%s", installRow.Productname, installRow.Rfidnum, err.Error())
		return false, err
	}
	var response []Response[interface{}]
	err = json.Unmarshal([]byte(*resultStr), &response)
	if err != nil {
		log.Println("反序列化构件安装结果信息失败", err.Error())
		return false, err
	}
	if len(response) == 0 || response[0].Code != "000" {
		return false, errors.New(response[0].Message)
	}
	return true, nil
}

// 使用ECB模式加密密码
func encryptPwd(password string) (string, error) {
	encryptBytes := []byte(password)
	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		return "", err
	}

	encryptBytes = myaes.Pkcs7Padding(encryptBytes, block.BlockSize())
	blockMode := myaes.NewECBEncrypter(block)
	encrypted := make([]byte, len(encryptBytes))
	blockMode.CryptBlocks(encrypted, encryptBytes)
	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// 使用ECB模式解密密码
func decryptPwd(password string) (string, error) {
	decryptBytes, _ := base64.StdEncoding.DecodeString(password)
	block, err := aes.NewCipher([]byte(aesKey))
	if err != nil {
		return "", err
	}

	blockMode := myaes.NewECBDecrypter(block)
	decrypted := make([]byte, len(decryptBytes))
	blockMode.CryptBlocks(decrypted, decryptBytes)
	decrypted = myaes.Pkcs7UnPadding(decrypted)
	return string(decrypted), nil
}
