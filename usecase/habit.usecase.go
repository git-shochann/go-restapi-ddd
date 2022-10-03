package usecase

import (
	"ddd/domain/logic"
	"ddd/domain/model"
	"ddd/domain/repository"
	"ddd/infrastructure/validator"
	"log"
	"net/http"
)

// habitの取得や登録などでDBにアクセスする時に、domain層のrepository(インターフェースとして設定した部分)を介してアクセスすることによって、infrastructure層にアクセスするのではなく、
// domain層のみに直接依存するだけで完結出来る！ 単一方向であるので。infrastructure層を触れたりすることはしない。

// ここのusecase層がすることは、図の上のinterface層から情報を受け取り、下のdomain層のインターフェースで定義してあるメソッドを用いてビジネスロジックを実行すること

// インターフェース -> 窓口である
type HabitUseCase interface {
	CreateHabit(h *model.Habit) error
	// DeleteHabit(habitID, userID int, habit *model.Habit) error
	// UpdateHabit(habit *model.Habit) error
	// GetAllHabitByUserID(user model.User, habit *[]model.Habit) error
}

// これはなに？ -> ここの層でやることを構造体で表現する。
// usecase層の関数をメソッドとして定義し、
// 構造体のフィールド内にインターフェース型として設定すれば、インターフェースはメソッドを使用することの出来る窓口であるので、
// いつでもそのメソッドを使うことが出来る
// その使うことが出来るメソッドは domain層のインターフェースで設定したインターフェース。

// どの方向に依存しているかで考えると分かりやすい。
type habitUseCase struct {
	HabitRepository      repository.HabitRepository //以下全てdomain層のインターフェース。 この構造体に紐づいているメソッドでそのメソッドを使用したいので！
	HabitValidation      validator.HabitValidation
	EncryptPassWordLogic logic.EncryptPasswordLogic
	EnvLogic             logic.EnvLogic
	JwtLogic             logic.JwtLogic
	LoggingLogic         logic.LoggingLogic
	ResponseLogic        logic.ResponseLogic
}

// インターフェースを引数にとってインターフェースを返す？ -> この引数はどこでそもそも呼び出す？
func NewHabitUseCase(hr repository.HabitRepository, hv validator.HabitValidation, epl logic.EncryptPasswordLogic, el logic.EnvLogic, jl logic.JwtLogic, ll logic.LoggingLogic, rl logic.ResponseLogic) HabitUseCase {
	return &habitUseCase{
		HabitRepository:      hr,
		HabitValidation:      hv,
		EncryptPassWordLogic: epl,
		JwtLogic:             jl,
		LoggingLogic:         ll,
		ResponseLogic:        rl,
	}
}

// domainのインターフェースを使って、実際に処理を行う
func (hu *habitUseCase) CreateHabit(habit *model.Habit) error {

	// 実際のDBの処理であるhu.CreateHabit() としてアクセスをすることが可能

	err := hu.CreateHabit(habit)
	if err != nil {
		hu.ResponseLogic.SendErrorResponse(w, "Failed to create habit", http.StatusInternalServerError)
		log.Println(err)
	}

	// response, err := json.Marshal(habit)
	// if err != nil {
	// 	models.SendErrorResponse(w, "Failed to read json", http.StatusBadRequest)
	// 	log.Println(err)
	// 	return
	// }

	// 	models.SendResponse(w, response, http.StatusOK)

	return nil

}

// 参考 //

// ということはこのCreateTodoを読んでいるところはどこ？

// func (ts *todoService) CreateTodo(w http.ResponseWriter, r *http.Request, userId int) (models.BaseTodoResponse, error) {
// 	// ioutil: ioに特化したパッケージ
// 	reqBody, _ := ioutil.ReadAll(r.Body)
// 	var mutationTodoRequest models.MutationTodoRequest
// 	if err := json.Unmarshal(reqBody, &mutationTodoRequest); err != nil {
// 		log.Fatal(err)
// 		errMessage := "リクエストパラメータを構造体へ変換処理でエラー発生"
// 		ts.rl.SendResponse(w, ts.rl.CreateErrorStringResponse(errMessage), http.StatusInternalServerError)
// 		return models.BaseTodoResponse{}, err
// 	}
// 	// バリデーション
// 	if err := ts.tv.MutationTodoValidate(mutationTodoRequest); err != nil {
// 		// バリデーションエラーのレスポンスを送信
// 		ts.rl.SendResponse(w, ts.rl.CreateErrorResponse(err), http.StatusBadRequest)
// 		return models.BaseTodoResponse{}, err
// 	}

// 	var todo models.Todo
// 	todo.Title = mutationTodoRequest.Title
// 	todo.Comment = mutationTodoRequest.Comment
// 	todo.UserId = userId

// 	// todoデータ新規登録処理
// 	if err := ts.tr.CreateTodo(&todo); err != nil {
// 		ts.rl.SendResponse(w, ts.rl.CreateErrorStringResponse("データ取得に失敗しました。"), http.StatusInternalServerError)
// 		return models.BaseTodoResponse{}, err
// 	}

// 	// 登録したtodoデータ取得処理
// 	if err := ts.tr.GetTodoLastByUserId(&todo, userId); err != nil {
// 		var errMessage string
// 		var statusCode int
// 		// https://gorm.io/ja_JP/docs/error_handling.html
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			statusCode = http.StatusBadRequest
// 			errMessage = "該当データは存在しません。"
// 		} else {
// 			statusCode = http.StatusInternalServerError
// 			errMessage = "データ取得に失敗しました。"
// 		}
// 		// エラーレスポンス送信
// 		ts.rl.SendResponse(w, ts.rl.CreateErrorStringResponse(errMessage), statusCode)
// 		return models.BaseTodoResponse{}, err
// 	}

// 	// レスポンス用の構造体に変換
// 	responseTodos := ts.tl.CreateTodoResponse(&todo)

// 	return responseTodos, nil
// }