package repository

import (
	"errors"
	"github.com/course-go-autumn-2026/lecture6-architecture/after/domain"
)

type PostgresListingRepository struct {
	id int64
	nextID int64
	db map[int64]*storedListing{}
}

func (r *PostgresListingRepository)Create(list Listing){
	r.nextID++
	r.id := nextID
	listing := &storedListing{
		ID:        id,
		Title:     list.Title,
		Price:     list.Price,
		PhotoURL:  list.PhotoURL,
		SellerID:  list.SellerID,
		Status:    "pending",
		CreatedAt: time.Now(),
	}
	r.db[id] = listing
}

func (r *PostgresListingRepository)Get(idlist int64) Listing{
	return r.db[idlist];
}

func (r *PostgresListingRepository)update(idlist int64, statuslist string, pricelist int64){
	r.db[idlist].status = statuslist
	r.db[idlist].price = pricelist
}