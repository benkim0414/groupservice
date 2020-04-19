package groupservice

import (
	"context"
	"io/ioutil"
	"strings"

	grouppb "github.com/benkim0414/groupservice/pb"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/oauth2/google"
	admin "google.golang.org/api/admin/directory/v1"
	"google.golang.org/api/option"
)

const (
	adminEmail = "gunwoo@gunwoo.org"
	domain     = "gunwoo.org"
)

type service struct {
	client *admin.Service
}

func New(ctx context.Context, serviceAccountFilePath string) (*service, error) {
	data, err := ioutil.ReadFile(serviceAccountFilePath)
	if err != nil {
		return nil, err
	}
	conf, err := google.JWTConfigFromJSON(
		data,
		admin.AdminDirectoryGroupScope,
		admin.AdminDirectoryGroupMemberScope,
	)
	if err != nil {
		return nil, err
	}
	conf.Subject = adminEmail
	ts := conf.TokenSource(ctx)
	client, err := admin.NewService(ctx, option.WithTokenSource(ts))
	if err != nil {
		return nil, err
	}
	return &service{client: client}, nil
}

func (svc *service) ListGroups(ctx context.Context, req *grouppb.ListGroupsRequest) (*grouppb.ListGroupsResponse, error) {
	res, err := svc.client.Groups.List().Domain(domain).Do()
	if err != nil {
		return &grouppb.ListGroupsResponse{}, err
	}
	groups := make([]*grouppb.Group, len(res.Groups))
	for i, group := range res.Groups {
		groups[i] = &grouppb.Group{
			Id:          group.Id,
			Name:        group.Name,
			Email:       group.Email,
			Description: group.Description,
		}
	}
	return &grouppb.ListGroupsResponse{
		Groups:        groups,
		NextPageToken: res.NextPageToken,
	}, nil
}

func (svc *service) GetGroup(ctx context.Context, req *grouppb.GetGroupRequest) (*grouppb.Group, error) {
	groupID := strings.Split(req.Name, "/")[1]
	group, err := svc.client.Groups.Get(groupID).Do()
	if err != nil {
		return &grouppb.Group{}, err
	}
	return &grouppb.Group{
		Id:          group.Id,
		Name:        group.Name,
		Email:       group.Email,
		Description: group.Description,
	}, nil
}

func (svc *service) CreateGroup(ctx context.Context, req *grouppb.CreateGroupRequest) (*grouppb.Group, error) {
	group, err := svc.client.Groups.Insert(&admin.Group{
		Name:        req.Group.Name,
		Email:       req.Group.Email,
		Description: req.Group.Description,
	}).Do()
	if err != nil {
		return &grouppb.Group{}, err
	}
	return &grouppb.Group{
		Id:          group.Id,
		Name:        group.Name,
		Email:       group.Email,
		Description: group.Description,
	}, nil
}

func (svc *service) DeleteGroup(ctx context.Context, req *grouppb.DeleteGroupRequest) (*empty.Empty, error) {
	groupID := strings.Split(req.Name, "/")[1]
	err := svc.client.Groups.Delete(groupID).Do()
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}

func (svc *service) ListMembers(ctx context.Context, req *grouppb.ListMembersRequest) (*grouppb.ListMembersResponse, error) {
	groupID := strings.Split(req.Parent, "/")[1]
	res, err := svc.client.Members.List(groupID).Do()
	if err != nil {
		return &grouppb.ListMembersResponse{}, err
	}
	members := make([]*grouppb.Member, len(res.Members))
	for i, member := range res.Members {
		members[i] = &grouppb.Member{
			Id:     member.Id,
			Email:  member.Email,
			Status: member.Status,
		}
	}
	return &grouppb.ListMembersResponse{
		Members:       members,
		NextPageToken: res.NextPageToken,
	}, nil
}

func (svc *service) GetMember(ctx context.Context, req *grouppb.GetMemberRequest) (*grouppb.Member, error) {
	ss := strings.Split(req.Name, "/")
	groupID, memberID := ss[1], ss[3]
	member, err := svc.client.Members.Get(groupID, memberID).Do()
	if err != nil {
		return &grouppb.Member{}, err
	}
	return &grouppb.Member{
		Id:     member.Id,
		Email:  member.Email,
		Status: member.Status,
	}, nil
}

func (svc *service) CreateMember(ctx context.Context, req *grouppb.CreateMemberRequest) (*grouppb.Member, error) {
	groupID := strings.Split(req.Parent, "/")[1]
	member, err := svc.client.Members.Insert(groupID, &admin.Member{
		Email: req.Member.Email,
	}).Do()
	if err != nil {
		return &grouppb.Member{}, err
	}
	return &grouppb.Member{
		Id:     member.Id,
		Email:  member.Email,
		Status: member.Status,
	}, nil
}

func (svc *service) DeleteMember(ctx context.Context, req *grouppb.DeleteMemberRequest) (*empty.Empty, error) {
	ss := strings.Split(req.Name, "/")
	groupID, memberID := ss[1], ss[3]
	err := svc.client.Members.Delete(groupID, memberID).Do()
	if err != nil {
		return &empty.Empty{}, err
	}
	return &empty.Empty{}, nil
}
