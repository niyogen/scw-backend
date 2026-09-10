package database

import (
	"log"
	"time"

	"delivery-backend/internal/models"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedData seeds initial data for services, workers, vouchers, and demo user
func SeedData(db *gorm.DB) error {
	// Seed Services
	var serviceCount int64
	db.Model(&models.ServiceItem{}).Count(&serviceCount)
	if serviceCount == 0 {
		services := []models.ServiceItem{
			{
				Code:                "srv_plumbing",
				Title:               "Plumbing",
				HighlightedSubtitle: "Master certified with zero-leak guarantee",
				Description:         "Pipe repair, leak detection, drainage & sanitary fixtures",
				IsNew:               true,
				IconType:            "plumbing",
				StartingPrice:       "500 ฿",
				BasePrice:           500.0,
				Category:            "household",
				Options: []models.ServiceOption{
					{Name: "General Pipe Inspection", Description: "Ultrasonic leak detection & pressure test", Price: 500, Unit: "visit"},
					{Name: "Faucet / Mixer Tap Replacement", Description: "New sanitary fitting installation", Price: 350, Unit: "point"},
					{Name: "Drain Unclogging", Description: "High-pressure jetter cleaning", Price: 800, Unit: "point"},
				},
			},
			{
				Code:                "srv_electricity",
				Title:               "Electricity Service",
				HighlightedSubtitle: "Licensed electrical engineer",
				Description:         "Wiring inspection, circuit breakers, lighting & appliance fix",
				IsNew:               true,
				IconType:            "electricity",
				StartingPrice:       "550 ฿",
				BasePrice:           550.0,
				Category:            "electrical",
				Options: []models.ServiceOption{
					{Name: "Short Circuit & Safety Inspection", Description: "Ground wire check and breaker diagnostics", Price: 550, Unit: "visit"},
					{Name: "Light Fixture Installation", Description: "Chandeliers, LED tracks, or spotlights", Price: 300, Unit: "point"},
					{Name: "EV Charger / Breaker Upgrade", Description: "High load circuit installation", Price: 1500, Unit: "point"},
				},
			},
			{
				Code:                "srv_building",
				Title:               "Building Work",
				HighlightedSubtitle: "Structural & masonry specialist",
				Description:         "Masonry, wall plastering, tiling & structural maintenance",
				IsNew:               true,
				IconType:            "building_work",
				StartingPrice:       "800 ฿",
				BasePrice:           800.0,
				Category:            "construction",
			},
			{
				Code:                "srv_cleaning",
				Title:               "Cleaning Service",
				HighlightedSubtitle: "Hospital-grade sanitization",
				Description:         "Deep residential, move-in sanitization & recurring maid care",
				IsNew:               false,
				IconType:            "cleaning",
				StartingPrice:       "450 ฿",
				BasePrice:           450.0,
				Category:            "cleaning",
				Options: []models.ServiceOption{
					{Name: "Wall-Mounted AC Cleaning (9,000 - 18,000 BTU)", Description: "Deep chemical coil wash, drain tray, anti-fungus spray", Price: 650, Unit: "unit"},
					{Name: "Ceiling Cassette AC Cleaning (18,000 - 36,000 BTU)", Description: "4-way cassette deep overhaul & high-pressure flush", Price: 1200, Unit: "unit"},
					{Name: "Condenser External Wash", Description: "High-pressure compressor fan & fin cleaning", Price: 300, Unit: "unit"},
					{Name: "Antibacterial UV Sanitization", Description: "Hospital grade sterilization treatment", Price: 250, Unit: "room"},
				},
			},
			{
				Code:                "srv_caretaking",
				Title:               "Care Taking",
				HighlightedSubtitle: "Certified nursing assistants",
				Description:         "Elderly assistance, patient companion & gentle home care",
				IsNew:               true,
				IconType:            "care_taking",
				StartingPrice:       "600 ฿",
				BasePrice:           600.0,
				Category:            "care",
			},
			{
				Code:                "srv_carpenter",
				Title:               "Carpenter Service",
				HighlightedSubtitle: "Bespoke woodwork & furniture repair",
				Description:         "Custom furniture repair, door fitting & bespoke woodwork",
				IsNew:               false,
				IconType:            "carpenter",
				StartingPrice:       "550 ฿",
				BasePrice:           550.0,
				Category:            "carpentry",
			},
			{
				Code:                "srv_cook",
				Title:               "Cook & Chef",
				HighlightedSubtitle: "Private gourmet chef",
				Description:         "Private home meal preparation, culinary catering & dinner party",
				IsNew:               true,
				IconType:            "cook",
				StartingPrice:       "650 ฿",
				BasePrice:           650.0,
				Category:            "culinary",
			},
			{
				Code:                "srv_other",
				Title:               "Other Services",
				HighlightedSubtitle: "General handyman & repairs",
				Description:         "General handyman tasks, locksmith, painting & repairs",
				IsNew:               false,
				IconType:            "other",
				StartingPrice:       "400 ฿",
				BasePrice:           400.0,
				Category:            "general",
			},
		}

		for _, s := range services {
			if err := db.Create(&s).Error; err != nil {
				log.Printf("Failed to seed service %s: %v", s.Title, err)
			}
		}
		log.Println("Services seeded successfully")
	}

	// Seed Workers
	var workerCount int64
	db.Model(&models.Worker{}).Count(&workerCount)
	if workerCount == 0 {
		workers := []models.Worker{
			{
				WorkerCode:       "wrk_1",
				Name:             "Somchai Prasert",
				AvatarURL:        "https://images.unsplash.com/photo-1534528741775-53994a69daeb?w=300&auto=format&fit=crop&q=80",
				Age:              34,
				ExperienceYears:  7,
				Rating:           4.95,
				ReviewsCount:     238,
				ServiceTypes:     []string{"plumbing", "other"},
				Specializations:  []string{"Master Plumbing Pro", "Pipe Leak Repair", "Sanitary Fitting", "Drainage Cleansing"},
				Phone:            "+66 82 459 9123",
				Email:            "somchai.prasert@servicepro.th",
				Bio:              "Master certified Plumber with over 7 years of expertise in luxury residential high-rises and private houses. Equipped with ultrasonic leak detection and high-pressure jetters.",
				CompletedJobs:    540,
				HourlyRate:       "500 ฿ / visit",
				RateValue:        500.0,
				IsVerified:       true,
				IsAvailableToday: true,
				Badges:           []string{"Master Plumber", "Verified ID", "Zero-Leak Guarantee", "Same-Day Service"},
				Reviews: []models.WorkerReview{
					{AuthorName: "Piyawat N.", Rating: 5.0, Comment: "Very professional, arrived 10 minutes early and fixed the kitchen leak seamlessly!"},
					{AuthorName: "Sarah Jenkins", Rating: 5.0, Comment: "Speaks great English, replaced our main water pressure valve cleanly. Highly recommended!"},
				},
			},
			{
				WorkerCode:       "wrk_2",
				Name:             "Chaiwat Rattanakul",
				AvatarURL:        "https://images.unsplash.com/photo-1506794778202-cad84cf45f1d?w=300&auto=format&fit=crop&q=80",
				Age:              36,
				ExperienceYears:  10,
				Rating:           4.97,
				ReviewsCount:     310,
				ServiceTypes:     []string{"electricity", "other"},
				Specializations:  []string{"Licensed Master Electrician", "Circuit Breaker Box", "EV Charger Install", "Smart Lighting"},
				Phone:            "+66 81 928 3344",
				Email:            "chaiwat.r@servicepro.th",
				Bio:              "Licensed Senior Electrical Engineer with 10 years experience in residential wiring, safety inspections, breaker upgrades, and premium smart home installations.",
				CompletedJobs:    840,
				HourlyRate:       "550 ฿ / hour",
				RateValue:        550.0,
				IsVerified:       true,
				IsAvailableToday: true,
				Badges:           []string{"Licensed Master", "10y Veteran", "Safety Certified", "Zero Hazard Guaranteed"},
				Reviews: []models.WorkerReview{
					{AuthorName: "Tavee S.", Rating: 5.0, Comment: "Detected an electrical ground short circuit within 15 minutes, solved it safely."},
				},
			},
			{
				WorkerCode:       "wrk_3",
				Name:             "Karn Srisawat",
				AvatarURL:        "https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=300&auto=format&fit=crop&q=80",
				Age:              42,
				ExperienceYears:  12,
				Rating:           4.96,
				ReviewsCount:     290,
				ServiceTypes:     []string{"building_work", "other"},
				Specializations:  []string{"Masonry & Brickwork", "Wall Plaster & Tiling", "Waterproof Coating", "Structural Renovation"},
				Phone:            "+66 84 551 2289",
				Email:            "karn.building@servicepro.th",
				Bio:              "Senior Building contractor and master mason with 12+ years experience in quality residential brickwork, bathroom tiling, and structural repairs.",
				CompletedJobs:    710,
				HourlyRate:       "800 ฿ / day",
				RateValue:        800.0,
				IsVerified:       true,
				IsAvailableToday: true,
				Badges:           []string{"Master Mason", "Structural Builder", "Precision Tiler", "Insured Work"},
				Reviews: []models.WorkerReview{
					{AuthorName: "Chavalit M.", Rating: 5.0, Comment: "Re-tiled our balcony and repaired the waterproof layer. Flawless finishing!"},
				},
			},
			{
				WorkerCode:       "wrk_4",
				Name:             "Siriporn Thanakorn",
				AvatarURL:        "https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=300&auto=format&fit=crop&q=80",
				Age:              38,
				ExperienceYears:  8,
				Rating:           4.98,
				ReviewsCount:     382,
				ServiceTypes:     []string{"cleaning", "other"},
				Specializations:  []string{"Luxury Residential Cleaning", "Deep Sanitization", "Move-In/Move-Out", "Eco-Friendly Care"},
				Phone:            "+66 86 312 9901",
				Email:            "siriporn.t@cleanpro.th",
				Bio:              "Senior housekeeping lead with 8 years catering to luxury penthouses and residential estates. Uses hospital-grade eco sanitizers and HEPA vacuums.",
				CompletedJobs:    620,
				HourlyRate:       "450 ฿ / session",
				RateValue:        450.0,
				IsVerified:       true,
				IsAvailableToday: true,
				Badges:           []string{"5-Star Elite", "Super Clean Lead", "Background Checked", "Pet Friendly"},
				Reviews: []models.WorkerReview{
					{AuthorName: "Marcus Lind", Rating: 5.0, Comment: "We use Siriporn every 2 weeks. House is always immaculately spotless and fragrant!"},
				},
			},
			{
				WorkerCode:       "wrk_5",
				Name:             "Ananya Kittisuk",
				AvatarURL:        "https://images.unsplash.com/photo-1580489944761-15a19d654956?w=300&auto=format&fit=crop&q=80",
				Age:              31,
				ExperienceYears:  6,
				Rating:           4.93,
				ReviewsCount:     165,
				ServiceTypes:     []string{"care_taking"},
				Specializations:  []string{"Elderly Companion Care", "Post-Op Assistance", "Mobility Support", "Medication Reminders"},
				Phone:            "+66 93 219 8830",
				Email:            "ananya.care@servicepro.th",
				Bio:              "Certified Nursing Assistant & geriatric caregiver with CPR/First Aid credentials. Gentle, patient, and trained in specialized compassionate home care.",
				CompletedJobs:    340,
				HourlyRate:       "600 ฿ / session",
				RateValue:        600.0,
				IsVerified:       true,
				IsAvailableToday: true,
				Badges:           []string{"CNA Certified", "CPR & First Aid", "Gentle Touch", "Verified Caregiver"},
				Reviews: []models.WorkerReview{
					{AuthorName: "Danielle W.", Rating: 5.0, Comment: "Looked after my mother with utmost kindness and patience. Truly heartwarming."},
				},
			},
			{
				WorkerCode:       "wrk_6",
				Name:             "Prasit Udomsak",
				AvatarURL:        "https://images.unsplash.com/photo-1500648767791-00dcc994a43e?w=300&auto=format&fit=crop&q=80",
				Age:              39,
				ExperienceYears:  11,
				Rating:           4.94,
				ReviewsCount:     220,
				ServiceTypes:     []string{"carpenter", "other"},
				Specializations:  []string{"Custom Cabinetry", "Door & Window Joinery", "Hardwood Flooring", "Furniture Restoration"},
				Phone:            "+66 89 772 1049",
				Email:            "prasit.carpentry@servicepro.th",
				Bio:              "Artisan woodworker and master carpenter specializing in solid teak repairs, bespoke cabinetry, door alignment, and custom wardrobe adjustments.",
				CompletedJobs:    490,
				HourlyRate:       "550 ฿ / hour",
				RateValue:        550.0,
				IsVerified:       true,
				IsAvailableToday: true,
				Badges:           []string{"Master Woodworker", "Precision Joinery", "Artisan Quality", "Insured Pro"},
				Reviews: []models.WorkerReview{
					{AuthorName: "Ploy Supaporn", Rating: 4.9, Comment: "Restored our vintage dining table and fixed sticking sliding doors perfectly."},
				},
			},
			{
				WorkerCode:       "wrk_7",
				Name:             "Ratree Phromma",
				AvatarURL:        "https://images.unsplash.com/photo-1573496359142-b8d87734a5a2?w=300&auto=format&fit=crop&q=80",
				Age:              35,
				ExperienceYears:  8,
				Rating:           4.99,
				ReviewsCount:     345,
				ServiceTypes:     []string{"cook"},
				Specializations:  []string{"Private Home Chef", "Authentic Thai Cuisine", "Western Gourmet", "Special Diet Prep"},
				Phone:            "+66 91 883 4567",
				Email:            "ratree.culinary@servicepro.th",
				Bio:              "Graduate of Culinary Arts with 8 years as private estate chef. Expert in balanced family nutrition, dinner party catering, and specialized dietary menus.",
				CompletedJobs:    580,
				HourlyRate:       "650 ฿ / meal",
				RateValue:        650.0,
				IsVerified:       true,
				IsAvailableToday: true,
				Badges:           []string{"Executive Chef", "Food Safety Cert", "5-Star Dining", "Gourmet Pro"},
				Reviews: []models.WorkerReview{
					{AuthorName: "Alexandre Roy", Rating: 5.0, Comment: "Prepared an extraordinary 4-course dinner for our family. Cleaned the kitchen spotlessly too!"},
				},
			},
		}

		for _, w := range workers {
			if err := db.Create(&w).Error; err != nil {
				log.Printf("Failed to seed worker %s: %v", w.Name, err)
			}
		}
		log.Println("Workers seeded successfully")
	}

	// Seed Vouchers
	var voucherCount int64
	db.Model(&models.Voucher{}).Count(&voucherCount)
	if voucherCount == 0 {
		vouchers := []models.Voucher{
			{
				Code:            "WELCOME20",
				Title:           "Welcome Discount",
				Description:     "20% off on your first home service order",
				DiscountPercent: 20.0,
				DiscountAmount:  0,
				MinSpend:        300,
				MaxDiscount:     500,
				ValidUntil:      time.Now().AddDate(1, 0, 0),
				IsActive:        true,
			},
			{
				Code:            "SAVE100",
				Title:           "Special Flat Discount",
				Description:     "Save 100 ฿ on any service above 500 ฿",
				DiscountPercent: 0,
				DiscountAmount:  100.0,
				MinSpend:        500,
				MaxDiscount:     100,
				ValidUntil:      time.Now().AddDate(1, 0, 0),
				IsActive:        true,
			},
		}

		for _, v := range vouchers {
			_ = db.Create(&v).Error
		}
		log.Println("Vouchers seeded successfully")
	}

	// Seed Demo User (password: Password123!)
	var userCount int64
	db.Model(&models.User{}).Count(&userCount)
	if userCount == 0 {
		hashed, _ := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)
		demoUser := models.User{
			Name:         "John Doe",
			Email:        "user@example.com",
			Phone:        "+66 81 234 5678",
			PasswordHash: string(hashed),
			Role:         "customer",
			AvatarURL:    "https://images.unsplash.com/photo-1535713875002-d1d0cf377fde?w=300&auto=format&fit=crop&q=80",
			Addresses: []models.UserLocation{
				{
					Title:     "The Base Sukhumvit 77",
					Subtitle:  "The Base Sukhumvit 77, On Nut Rd, Phra Khanong Nuea, Watthana, Bangkok 10110",
					Latitude:  13.7142,
					Longitude: 100.6015,
					IsDefault: true,
				},
			},
		}
		_ = db.Create(&demoUser).Error
		log.Println("Demo user seeded successfully (user@example.com / Password123!)")
	}

	return nil
}
