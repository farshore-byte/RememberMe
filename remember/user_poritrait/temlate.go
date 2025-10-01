package user_poritrait

// 提示词建议使用redis存储，前期写在代码里调试
import (
	"fmt"
	"strings"
)

// -------------------------- 基础模版 ---------------------------

var UserProfilePromptTemplate = `
## Role
You are an expert at organizing and updating user profiles (memories).
Your task is to refine a user profile based on "Existing Memories" to reflect the most relevant and up-to-date information about the user.
Note: Users are often referred to simply as "Users."

## Task Description
"Existing Memories" are summaries of conversations between the protagonist, the user, and their assistants in a role-playing game.
You are required to accurately identify relevant information about the user in "New Conversations" and organize it into a more complete, accurate, and up-to-date user profile.

## Task Restrictions
- If new information conflicts with existing information, carefully evaluate:
- If the new conversation provides updated or more accurate information, revise the existing profile.
- If the new information is ambiguous and not explicitly confirmed by the user, retain the original information and discard the new information.
- If the new information does not appear in "Existing Memories," append it to the existing entry, separated by a semicolon.
- All entries should use a consistent and clear writing style. Ensure that each entry is well-documented and clearly sourced.
- If the extracted information contains time information, the actual date must be inferred from the current time (e.g., "last Wednesday," "yesterday," "tomorrow") and converted to a specific date.
- Output the most valuable and information-dense fields, ignoring fields not explicitly specified in the conversation.
- Pay attention to the nouns related to user information used in the original text

## User Profile Overview
1. User Basic Information Field Name: {basic_information_list}
2. User Followed Topics Field Name: {topic_list}
3. User Sexual Orientation Field Name: {sexual_list}

## User Profile Description Requirements
1. Language should be descriptive. The output will be displayed directly to end users, so keep the language natural, fluent, and rich.
2. Language should be as rich and complete as possible. Avoid simply listing keywords. Sentences do not need to contain a topic.

## Profile
## Existing Memories

{current_user_portrait}

## New Conversation
### Current Conversation History

{messages_str}

# Current Time
{current_time}

# Output Format Example
The output must be in JSON format, with the key being the field name and the value being the updated/merged information.

Note: If the merged value is empty or None, the field is skipped and not output.

{format_example}

# Output Example 1

{output_example_1}

# Output Example 2

{output_example_2}

# Output Example 3

{output_example_3}

# Language Settings
- Your working language is only {language}

# If no updates were made, the update JSON result will be {}

`

// ------------------------- 业务变量 ---------------------------------

// 基本信息
var BasicInformationList = []string{
	"name",             // 姓名，真实身份标识
	"nickname",         // 昵称，社交/日常使用
	"language",         // 语言/母语，便于多语言适配
	"gender",           // 性别
	"citizenship",      // 国籍
	"birthday",         // 生日（可推算年龄、星座）
	"height",           // 身高
	"weight",           // 体重
	"age",              // 年龄
	"hometown",         // 籍贯/家乡
	"residence",        // 居住地（现居城市）
	"identity",         // 身份（学生、上班族、自由职业等）
	"occupation",       // 职业
	"relationships",    // 人际关系（单身/恋爱/已婚/亲子）
	"marital_status",   // 婚姻状态
	"contact",          // 联系方式（手机号/邮箱/社交账号）
	"education",        // 教育程度（高中、本科、硕士…）
	"income",           // 收入水平（可用于消费能力分层）
	"personality",      // 性格特征（内向/外向/开放）
	"habits",           // 习惯（作息、饮食、运动）
	"mood",             // 情绪倾向（积极/焦虑/悲观）
	"health",           // 健康状况（体质、疾病史）
	"values",           // 价值观（传统、现代、创新）
	"beliefs",          // 信仰（宗教、哲学观）
	"skills",           // 技能特长
	"digital_behavior", // 数字行为（常用设备、上网习惯）
	"interests",        // 兴趣
	"preferences",      // 偏好
}

// 兴趣话题
var TopicList = []string{
	"celebrities",          // 名人、明星
	"public_figures",       // 政治人物、社会名人
	"fictional_characters", // 动漫、小说、影视中的角色
	"literature",           // 文学、书籍
	"electronics",          // 电子产品（手机、电脑、智能设备）
	"games",                // 游戏（手游、端游、桌游）
	"sports",               // 体育（足球、篮球、网球…）
	"outdoor_activities",   // 户外活动（露营、爬山、骑行）
	"movies_tv",            // 影视剧、综艺
	"food",                 // 美食（烹饪、餐厅、饮食文化）
	"pets",                 // 宠物（猫、狗、异宠）
	"travel",               // 旅游（国内/国外）
	"education",            // 教育、学习（外语、学科、课程）
	"technology",           // 科技（AI、互联网、科研）
	"workplace",            // 职场话题（职业发展、公司文化）
	"cars",                 // 汽车（新能源、改装、赛车）
	"fashion",              // 时尚（穿搭、美妆、潮流）
	"housing",              // 房产、租房、装修
	"transportation",       // 出行方式（公交、地铁、出行服务）
	"finance",              // 金融（投资、理财、股票、加密货币）
	"music",                // 音乐（流行、古典、乐器）
	"art",                  // 艺术（绘画、雕塑、设计）
	"history",              // 历史、文化
	"philosophy",           // 哲学、思考方式
	"health_fitness",       // 健康、健身、养生
	"relationships",        // 两性关系、友情、亲情
}

// 性倾向
var SexualList = []string{
	"sexual_orientation // 性取向（异性恋、同性恋、双性恋等）",
	"sexual_intercourse_type // 行为类型（如口交、后入式、乳交）输出sexual_intercourse_type字段之后必须优先输出body_part",
	"body_part // 身体部位，从sexual_intercourse_type中提取：嘴、胸、屁股、阴唇、乳头、足、手.....",
	"scene // 场景类型（卧室、大街、马圈、海滩...）",
	"props // 道具相关(跳蛋、项圈、自慰棒。。。。。。)",
	"sm_tendency // SM 倾向（轻度/重度/无）",
	"plot // 剧情偏好（强奸/乱伦/纯爱/师生）",
	"sexual_fetish // 恋物癖相关",
	"safety_preference // 安全偏好（安全套/内射)",
	"openness  // 开放程度（保守/中立/开放）",
	"partner_preference// 偏好的伴侣特征（年龄段、性格、身份、职业等）",
}

// 需求
var RequirementList = []string{
	"physiological_needs",
	"social_needs",
	"safety_needs",
	"esteem_needs",
	"self_actualization_needs",
}

// 格式示例
var UserProfileFormatExample = `The user's personal information fields that need to be updated and are not empty include ["gender","music"].

The following are the incremental updates:
- basic_information:
{"name":"xxx","gender":"xxx","language":"xxx"}

- interest_topics:
{"music":"xxx","food":"xxx"}

`

// 示例1
var UserProfileOutputExample1 = `The user's personal information fields that need to be updated and are not empty include ["name","gender","language"].

The following are the incremental updates:
- basic_information:
    {"name":"David","gender":"Male","language":"English"}

- interest_topics:
    {"music":"Listens to ambient electronic and lo-fi while coding and writing; enjoys indie rock at night to relieve stress"}

- sexual_orientation:
    {"body_part":"mouth;tit;hand;","sexual_intercourse_type":"Oral sex; titjob; straddling;","Plot":"Infidelity; training; outdoor exposure;","Oral organs":"Big ass; deep throat; big breasts;","Partner character":"Wife, neighbor's wife;"}
`

// 示例二
var UserProfileOutputExample2 = `The user's personal information fields that need to be updated and are not empty include ["name","gender","identity","residence","food","travel"].

The following are the incremental updates: 

- basic_information:
    {"name":"Nancy","gender":"Female","identity":"Manager","residence":"USA"}

- interest_topics:
    {"food":"Enjoys strawberry and matcha cakes at a local bakery",
     "travel":"Likes Southern Europe and has visited Serbia once"}

`

// 示例三
var UserProfileOutputExample3 = `The user's personal information fields that need to be updated and are not empty include  ["name","Personality","habits"].

The following are the incremental updates: 

- basic_information:
    {"name":"Anna","personality":"Patient; Kind;","habits":"Very tidy, enjoys organizing things"}

- interest_topics:
    {}

`

// 工作语言设置
var UserProfilelanguage = "English"

// ------------------------- 代码 -----------------------------

var UserProfileStaticVars = map[string]string{
	"basic_information_list": strings.Join(BasicInformationList, ", "),
	"topic_list":             strings.Join(TopicList, ", "),
	"sexual_list":            strings.Join(SexualList, ", "),
	//"requirement_list":       strings.Join(RequirementList, ", "),
	"format_example":   UserProfileFormatExample,
	"output_example_1": UserProfileOutputExample1,
	"output_example_2": UserProfileOutputExample2,
	"output_example_3": UserProfileOutputExample3,
	"language":         UserProfilelanguage,
}

type UserProfileDynamicVars struct {
	CurrentUserPortrait string // 当前用户画像
	MesssagesStr        string // 当前对话记录
	CurrentTime         string // 当前时间
}

type UserProfileTemplate struct {
	Template   string            // 静态变量已替换后的模板
	StaticVars map[string]string // 静态变量
}

// ---------------------------
// 构造函数
// ---------------------------

var Template *UserProfileTemplate

func init() {
	Template = NewUserProfileTemplate()
}

func NewUserProfileTemplate() *UserProfileTemplate {
	return &UserProfileTemplate{
		Template:   UserProfilePromptTemplate,
		StaticVars: UserProfileStaticVars,
	}
}

func (u *UserProfileTemplate) BuildPrompt(dynamicVars *UserProfileDynamicVars) (string, error) {
	// 先替换静态变量
	finalTpl, err := SystemPromptComposeStatic(u.Template, u.StaticVars)
	if err != nil {
		return "", fmt.Errorf("静态模板组装失败: %v", err)
	}

	// 转成map[string]string 供systempromptCompose使用

	dynMap := map[string]string{
		"current_user_portrait": dynamicVars.CurrentUserPortrait,
		"messages_str":          dynamicVars.MesssagesStr,
		"current_time":          dynamicVars.CurrentTime,
	}

	// 再替换动态变量
	finalTpl, err = SystemPromptCompose(finalTpl, dynMap)
	if err != nil {
		return "", fmt.Errorf("动态模板组装失败: %v", err)
	}

	return finalTpl, nil
}
