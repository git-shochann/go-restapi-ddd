// interface層 (usecase層に依存)

package handler

import (
	"ddd/domain/model"
	"ddd/infrastructure"
	"ddd/interface/customerr"
	"ddd/interface/util"
	"ddd/interface/validator"
	"ddd/usecase"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
)

// ここの層に依存する箇所で使用する メソッドの窓口を用意してあげる
type HabitHandler interface {
	IndexFunc(w http.ResponseWriter, r *http.Request)
	CreateFunc(w http.ResponseWriter, r *http.Request)
	UpdateFunc(w http.ResponseWriter, r *http.Request)
	DeleteFunc(w http.ResponseWriter, r *http.Request)
	GetAllHabitFunc(w http.ResponseWriter, r *http.Request)
}

type habitHandler struct {
	HabitUseCase    usecase.HabitUseCase     // usecase層
	HabitValidation validator.HabitValidator // interface層
	JwtUtil         util.JwtUtil             // interface層
	ResponseUtil    util.ResponseUtil        // interface層
}

// ここの構造体のフィールドに書くのは、依存先のインターフェースを書けばOK
func NewHabitHandler(habitUseCase usecase.HabitUseCase, habitValidation validator.HabitValidator, jwtUtil util.JwtUtil, ResponseUtil util.ResponseUtil) HabitHandler {
	return &habitHandler{
		HabitUseCase:    habitUseCase,
		HabitValidation: habitValidation,
		JwtUtil:         jwtUtil,
		ResponseUtil:    ResponseUtil,
	}
}

// 第一引数にはHTTPサーバーからのレスポンスを出力することが出来るメソッドを持っている(該当のメソッドを実装している)構造体の値が来る
func (hh *habitHandler) IndexFunc(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("r.Body: %v\n", r.Body)
	fmt.Printf("%T\n", w)                   // *http.response構造体
	fmt.Fprintf(w, "this is Go's Rest API") // メソッド内でw.Write()をするため
}

// main（）のrouter.HandleFunc()の第二引数として以下の関数を渡すだけ
func (hh *habitHandler) CreateFunc(w http.ResponseWriter, r *http.Request) {

	// JWTの検証
	userID, err := hh.JwtUtil.CheckJWTToken(r)
	if err != nil {

		log.Println(err)

		var jwtErr *customerr.JwtErr

		switch {
		// error型の変数を引数に取る
		case errors.Is(err, customerr.ErrInvalidToken):
			hh.ResponseUtil.SendErrorResponse(w, "invalid token", http.StatusBadRequest)
		case errors.Is(err, customerr.ErrInvalidSignature):
			hh.ResponseUtil.SendErrorResponse(w, "invalid token", http.StatusBadRequest)
		case errors.Is(err, customerr.ErrAssertType):
			hh.ResponseUtil.SendErrorResponse(w, "invalid token", http.StatusBadRequest)
		case errors.As(err, &jwtErr): // errの元の型が *jwtErrだった場合
			hh.ResponseUtil.SendErrorResponse(w, "jwt error", http.StatusBadRequest)
		default:
			hh.ResponseUtil.SendErrorResponse(w, "unknown error occured", http.StatusInternalServerError)
		}
		return
	}

	// Bodyの読み込み
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		hh.ResponseUtil.SendErrorResponse(w, "failed to read json", http.StatusBadRequest)
		return // router.HandleFunc())の第二引数に関数を渡すだけなので戻り値なし
	}

	// バリデーション
	var habitValidation model.CreateHabitValidation
	err = json.Unmarshal(reqBody, &habitValidation)
	if err != nil {
		log.Println(err)
		hh.ResponseUtil.SendErrorResponse(w, "failed to read json", http.StatusBadRequest)
		return
	}

	errorMessage, err := hh.HabitValidation.HabitValidate(&habitValidation)
	if err != nil {
		log.Println(err)
		hh.ResponseUtil.SendErrorResponse(w, errorMessage, http.StatusBadRequest)
		return
	}

	// DBに登録する内容の準備
	habit := model.Habit{
		Content:  habitValidation.Content,
		Finished: false,
		UserID:   userID,
	}

	// 保存処理
	newHabit, err := hh.HabitUseCase.CreateHabit(&habit) // -> usecase層に依存
	if err != nil {
		log.Println(err)

		// エラーは2種類
		// ErrRecordNotFound
		// DbErr

		var DbErr *infrastructure.DbErr

		switch {
		case errors.Is(err, infrastructure.ErrRecordNotFound):
			hh.ResponseUtil.SendErrorResponse(w, "record not found", http.StatusBadRequest)
		case errors.As(err, &DbErr):
			hh.ResponseUtil.SendErrorResponse(w, "failed to create habit", http.StatusBadRequest)
		default:
			hh.ResponseUtil.SendErrorResponse(w, "unknown error occured", http.StatusInternalServerError)
		}

		return
	}

	// 登録が完了したhabitを上書きしてレスポンスとして返すためにjson形式にする([]byte)
	response, err := json.Marshal(newHabit)
	if err != nil {
		log.Println(err)
		hh.ResponseUtil.SendErrorResponse(w, "failed to encode json", http.StatusBadRequest)
		return
	}

	hh.ResponseUtil.SendResponse(w, response, http.StatusOK)

}

// WIP: 現在1つのIDを送ってるのにそのユーザーに紐付く習慣全て変わってる
func (hh *habitHandler) UpdateFunc(w http.ResponseWriter, r *http.Request) {

	// JWTの検証
	userID, err := hh.JwtUtil.CheckJWTToken(r)
	if err != nil {
		log.Println(err)
		var jwtErr *customerr.JwtErr

		switch {
		// error型の変数を引数に取る
		case errors.Is(err, customerr.ErrInvalidToken):
			hh.ResponseUtil.SendErrorResponse(w, "invalid token", http.StatusBadRequest)
		case errors.Is(err, customerr.ErrInvalidSignature):
			hh.ResponseUtil.SendErrorResponse(w, "invalid token", http.StatusBadRequest)
		case errors.Is(err, customerr.ErrAssertType):
			hh.ResponseUtil.SendErrorResponse(w, "invalid token", http.StatusBadRequest)
		case errors.As(err, &jwtErr):
			hh.ResponseUtil.SendErrorResponse(w, "jwt error", http.StatusBadRequest)
		default:
			hh.ResponseUtil.SendErrorResponse(w, "unknown error occured", http.StatusInternalServerError)
		}

		return
	}

	// 確認したJWTのクレームのuser_id
	// パスパラメーターから取得する habitのid

	vars := mux.Vars(r)
	// fmt.Printf("vars: %v\n", vars) // vars: map[id:1]
	habitIDStr := vars["id"]

	habitID, err := strconv.Atoi(habitIDStr)
	if err != nil {
		hh.ResponseUtil.SendErrorResponse(w, "something wrong", http.StatusBadRequest)
		return
	}

	// Bodyを検証
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		hh.ResponseUtil.SendErrorResponse(w, "failed to read json", http.StatusBadRequest)
		return
	}

	// バリデーションの実施
	var habitValidation model.CreateHabitValidation
	err = json.Unmarshal(reqBody, &habitValidation)
	if err != nil {
		hh.ResponseUtil.SendErrorResponse(w, "failed to read json", http.StatusBadRequest)
		return
	}

	errorMessage, err := hh.HabitValidation.HabitValidate(&habitValidation)
	if err != nil {
		hh.ResponseUtil.SendErrorResponse(w, errorMessage, http.StatusBadRequest)
		return
	}

	habit := model.Habit{
		Model: gorm.Model{
			ID: uint(habitID),
		}, // habitテーブルのid
		Content: habitValidation.Content, // content
		UserID:  userID,                  // user_id
	}

	updatedHabit, err := hh.HabitUseCase.UpdateHabit(&habit)

	if err != nil {
		log.Println(err)

		var DbErr *infrastructure.DbErr

		switch {
		case errors.Is(err, infrastructure.ErrRecordNotFound):
			hh.ResponseUtil.SendErrorResponse(w, "record not found", http.StatusBadRequest)
		case errors.As(err, &DbErr):
			hh.ResponseUtil.SendErrorResponse(w, "failed to update habit", http.StatusBadRequest)
		default:
			hh.ResponseUtil.SendErrorResponse(w, "unknown error occured", http.StatusInternalServerError)
		}

		return
	}

	response, err := json.Marshal(updatedHabit)
	if err != nil {
		hh.ResponseUtil.SendErrorResponse(w, "failed to encode json", http.StatusBadRequest)
		return
	}

	hh.ResponseUtil.SendResponse(w, response, http.StatusOK)
}

func (hh *habitHandler) DeleteFunc(w http.ResponseWriter, r *http.Request) {

	// JWTの検証
	userID, err := hh.JwtUtil.CheckJWTToken(r)
	if err != nil {
		log.Println(err)
		var jwtErr *customerr.JwtErr

		switch {
		// error型の変数を引数に取る
		case errors.Is(err, customerr.ErrInvalidToken):
			hh.ResponseUtil.SendErrorResponse(w, "invalid token", http.StatusBadRequest)
		case errors.Is(err, customerr.ErrInvalidSignature):
			hh.ResponseUtil.SendErrorResponse(w, "invalid token", http.StatusBadRequest)
		case errors.Is(err, customerr.ErrAssertType):
			hh.ResponseUtil.SendErrorResponse(w, "invalid token", http.StatusBadRequest)
		case errors.As(err, &jwtErr):
			hh.ResponseUtil.SendErrorResponse(w, "jwt error", http.StatusBadRequest)
		default:
			hh.ResponseUtil.SendErrorResponse(w, "unknown error occured", http.StatusInternalServerError)
		}

		return
	}

	// 確認したJWTのクレームのuser_id + パスパラメーターから取得する habitのidで削除処理を実装する
	vars := mux.Vars(r)
	fmt.Printf("vars: %v\n", vars) // vars: map[id:1]
	habitIDStr := vars["id"]

	habitID, err := strconv.Atoi(habitIDStr)
	if err != nil {
		hh.ResponseUtil.SendErrorResponse(w, "something wrong", http.StatusBadRequest)
		return
	}

	var habit model.Habit

	err = hh.HabitUseCase.DeleteHabit(habitID, userID, &habit)
	if err != nil {
		log.Println(err)

		var DbErr *infrastructure.DbErr

		switch {
		case errors.Is(err, infrastructure.ErrRecordNotFound):
			hh.ResponseUtil.SendErrorResponse(w, "recordnot found", http.StatusBadRequest)
		case errors.As(err, &DbErr):
			hh.ResponseUtil.SendErrorResponse(w, "failed to delete habit", http.StatusBadRequest)
		default:
			hh.ResponseUtil.SendErrorResponse(w, "unknown error occured", http.StatusInternalServerError)
		}

		return
	}

	hh.ResponseUtil.SendResponse(w, nil, http.StatusOK)

}

// ユーザー1人が持っているhabitを全て取得する
func (hh *habitHandler) GetAllHabitFunc(w http.ResponseWriter, r *http.Request) {

	// JWTの検証
	userID, err := hh.JwtUtil.CheckJWTToken(r)
	if err != nil {
		log.Println(err)
		var jwtErr *customerr.JwtErr

		switch {
		// error型の変数を引数に取る
		case errors.Is(err, customerr.ErrInvalidToken):
			hh.ResponseUtil.SendErrorResponse(w, "invalid token", http.StatusBadRequest)
		case errors.Is(err, customerr.ErrInvalidSignature):
			hh.ResponseUtil.SendErrorResponse(w, "invalid token", http.StatusBadRequest)
		case errors.Is(err, customerr.ErrAssertType):
			hh.ResponseUtil.SendErrorResponse(w, "invalid token", http.StatusBadRequest)
		case errors.As(err, &jwtErr):
			hh.ResponseUtil.SendErrorResponse(w, "jwt error", http.StatusBadRequest)
		default:
			hh.ResponseUtil.SendErrorResponse(w, "unknown error occured", http.StatusInternalServerError)
		}

		return
	}

	user := model.User{
		Model: gorm.Model{
			ID: uint(userID),
		},
	}

	var habit []model.Habit
	allHabit, err := hh.HabitUseCase.GetAllHabitByUserID(&user, &habit) // 旧: 値を渡す, 新: ポインタ(アドレス)を渡すことでしっかりと返却された
	if err != nil {
		log.Println(err)

		var DbErr *infrastructure.DbErr

		switch {
		case errors.Is(err, infrastructure.ErrRecordNotFound):
			hh.ResponseUtil.SendErrorResponse(w, "not found record", http.StatusBadRequest)
		case errors.As(err, &DbErr):
			hh.ResponseUtil.SendErrorResponse(w, "failed to get all habit", http.StatusBadRequest)
		default:
			hh.ResponseUtil.SendErrorResponse(w, "unknown error occured", http.StatusInternalServerError)
		}

		return
	}

	response, err := json.Marshal(allHabit)
	if err != nil {
		hh.ResponseUtil.SendErrorResponse(w, "failed to read json", http.StatusBadRequest)
		return
	}

	hh.ResponseUtil.SendResponse(w, response, http.StatusOK)
}
