package service

import (
	"github.com/course-go-autumn-2026/lecture6-architecture/after/domain"
	"github.com/course-go-autumn-2026/lecture6-architecture/after/repository"
)

type ListingService struct {
	repo ListingRepository
	moderator Moderator
}

func NewListingService(repo ListingRepository, moderator Moderator) *ListingService {
	return &ListingService{repo: repo, moderator: moderator}
}

type PublishInput struct {
	Title string
	Price float64
	PhotoURL string
	SellerID int64
}

func (r *ListingService) findByID(id int64) Listing{
	return r.repo.Get(id)
}

func (r *ListingService) updateListing(id int64, status string, price int64){
	r.repo.update(id, status, price)
}