package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/byebyebymyai/oauth2-api/endpoint"
	httpTransport "github.com/byebyebymyai/oauth2-api/transport/http"
	"github.com/google/uuid"
)

type User struct {
	// ID of the ent.
	ID *uuid.UUID `json:"id,omitempty"`
	// 工号
	Username string `json:"username,omitempty"`
	// 密码
	Password []byte `json:"password,omitempty"`
	// 是否管理员
	IsAdmin *bool `json:"is_admin,omitempty"`
	// 员工类型
	Type *string `json:"type,omitempty"`
	// 姓名
	Name *string `json:"name,omitempty"`
	// 证件类型
	PersonalIDType *string `json:"personal_id_type,omitempty"`
	// 证件号码
	PersonalIDNumber *string `json:"personal_id_number,omitempty"`
	// 国籍
	NationalityID *uuid.UUID `json:"nationality_id,omitempty"`
	// 联系电话
	Phone *string `json:"phone,omitempty"`
	// 批复日期
	ApprovalDate *time.Time `json:"approval_date,omitempty"`
	// 任职日期
	EntryDate *time.Time `json:"entry_date,omitempty"`
	// 员工状态
	UserStatus *string `json:"user_status,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the UserQuery when eager-loading is set.
	Edges      *UserEdges    `json:"edges"`
	Roles      []*Role       `json:"roles,omitempty"`
	Permission []*Permission `json:"permission,omitempty"`
}

type UserEdges struct {
	// Roles holds the value of the roles edge.
	Roles []*Role `json:"roles,omitempty"`
	// Group holds the value of the group edge.
	Group *Group `json:"group,omitempty"`
	// Position holds the value of the position edge.
	Position *Position `json:"position,omitempty"`
}

type Role struct {
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// 名称
	Name string `json:"name,omitempty"`
	// Edges holds the relations/edges for other nodes in the graph.
	// The values are being populated by the RoleQuery when eager-loading is set.
	Edges RoleEdges `json:"edges"`
}

type RoleEdges struct {
	// Permissions holds the value of the permissions edge.
	Permissions []*Permission `json:"permissions,omitempty"`
}

type Permission struct {
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// 名称
	Name string `json:"name,omitempty"`
	// 协议
	Protocol string `json:"protocol,omitempty"`
	// 主机
	Host string `json:"host,omitempty"`
	// 端口
	Port string `json:"port,omitempty"`
	// 方法
	Method string `json:"method,omitempty"`
	// 路径
	Path string `json:"path,omitempty"`
}

type Group struct {
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// 银行机构代码
	Code string `json:"code,omitempty"`
	// 银行机构名称
	Name string `json:"name,omitempty"`
	// 内部机构号
	InternalCode *string `json:"internal_code,omitempty"`
	// 机构类别:管理机构;营业机构;虚拟机构;内设机构;
	Type *string `json:"type,omitempty"`
	// 企业证件类型
	CorporateIDType *string `json:"corporate_id_type,omitempty"`
	// 企业证件号码
	CorporateIDNumber *string `json:"corporate_id_number,omitempty"`
	// 行政区划代码
	RegionID *uuid.UUID `json:"region_id,omitempty"`
	// 营业状态
	GroupStatus *string `json:"group_status,omitempty"`
	// 机构地址
	Address *string `json:"address,omitempty"`
	// 机构联系电话
	Phone *string `json:"phone,omitempty"`
}

type Position struct {
	// ID of the ent.
	ID uuid.UUID `json:"id,omitempty"`
	// 岗位代码
	Code string `json:"code,omitempty"`
	// 岗位名称
	Name string `json:"name,omitempty"`
	// 岗位类型
	Type *string `json:"type,omitempty"`
}

func proxyTokenEndpoint(_ context.Context, instance string) endpoint.Endpoint {
	u, err := url.Parse(instance + "/jwt")
	if err != nil {
		panic(err)
	}
	return httpTransport.NewClient(
		http.MethodPost,
		u,
		httpTransport.EncodeJSONRequest,
		decodeTokenResponse,
		httpTransport.ClientBefore(httpTransport.PopulateRequestContext),
	).Endpoint()
}

func decodeTokenResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		text, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(text))
	}
	var result []string
	for _, v := range r.Header.Values("Authorization") {
		if strings.HasPrefix(v, "Bearer ") {
			result = append(result, strings.TrimPrefix(v, "Bearer "))
		}
	}
	return result, nil
}

type tokenGenerationRequest struct {
	Iss string   `json:"iss,omitempty"`
	Sub string   `json:"sub,omitempty"`
	Exp int64    `json:"exp,omitempty"`
	Aud []string `json:"aud,omitempty"`
}

func proxyUserAllEndpoint(_ context.Context, instance string) endpoint.Endpoint {
	u, err := url.Parse(instance + "/user")
	if err != nil {
		panic(err)
	}
	return httpTransport.NewClient(
		http.MethodGet,
		u,
		encodeUserSearchRequest,
		decodeUserListResponse,
		httpTransport.ClientBefore(httpTransport.PopulateRequestContext),
	).Endpoint()
}

func encodeUserSearchRequest(_ context.Context, r *http.Request, request interface{}) error {
	q := r.URL.Query()
	q.Add("username", request.(User).Username)
	r.URL.RawQuery = q.Encode()
	return nil
}

func decodeUserListResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		text, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(text))
	}
	var result []User
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}

func proxyUserGetEndpoint(_ context.Context, instance string) endpoint.Endpoint {
	u, err := url.Parse(instance + "/user/{userID}")
	if err != nil {
		panic(err)
	}
	return httpTransport.NewClient(
		http.MethodPost,
		u,
		encodeUserGetRequest,
		decodeUserResponse,
		httpTransport.ClientBefore(httpTransport.PopulateRequestContext),
	).Endpoint()
}

func encodeUserGetRequest(_ context.Context, r *http.Request, request interface{}) error {
	r.URL.Path = strings.Replace(r.URL.Path, "{userID}", request.(uuid.UUID).String(), 1)
	return nil
}

func decodeUserResponse(_ context.Context, r *http.Response) (interface{}, error) {
	if r.StatusCode != http.StatusOK {
		text, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		return nil, errors.New(string(text))
	}
	var result User
	if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}
