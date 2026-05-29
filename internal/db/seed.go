package db

import (
	"fmt"
	"time"

	"adediiji.uk/jmcann-suffolk-backpack-task/internal/model"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
)

func Seed(app *AppStorage, secret string) error {
	gofakeit.Seed(0)

	fmt.Println("Seeding users and operatives...")
	qsID, err := app.NewUser("Admin QS", "qs@backpack.dev", "password123", model.UserRoleQS)
	if err != nil {
		return fmt.Errorf("seed QS user: %w", err)
	}
	fmt.Printf("QS user created: qs@backpack.dev / password123 (id: %s)\n", qsID)

	operativeIDs := make([]*uuid.UUID, 0)
	operativeNames := []string{"John Hamm", "Sarah Okafor", "Marcus Webb", "Diane Fletcher", "Ray Kowalski"}
	operativeTrades := []string{"Electrician", "Traffic Management", "Civil Engineer", "Street Lighting Engineer", "Cable Jointer"}
	for i := 0; i < 5; i++ {
		email := fmt.Sprintf("%s@backpack.dev", gofakeit.Username())
		phone := gofakeit.Phone()
		operativeUserID, err := app.NewUser(operativeNames[i], email, "password123", model.UserRoleOperative)
		if err != nil {
			return fmt.Errorf("seed operative user: %w", err)
		}
		operativeID, err := app.NewOperative(operativeUserID, operativeNames[i], email, &phone, &operativeTrades[i])
		if err != nil {
			return fmt.Errorf("seed operative: %w", err)
		}
		rate := model.NewMoney("GBP", uint64(gofakeit.IntRange(1500, 3500)))
		_, err = app.NewOperativeRate(operativeID, rate, time.Now())
		if err != nil {
			return fmt.Errorf("seed operative rate: %w", err)
		}
		operativeIDs = append(operativeIDs, operativeID)
		fmt.Printf("Operative created: %s (%s) rate: %s\n", operativeNames[i], email, rate.Serialize())
	}

	fmt.Println("Seeding sites...")
	siteIDs := make([]*uuid.UUID, 0)
	siteNames := []string{
		fmt.Sprintf("Woodbridge Road Zone %d", 1),
		fmt.Sprintf("Ipswich Central Zone %d", 2),
		fmt.Sprintf("Felixstowe Dock Road Zone %d", 3),
	}
	for i := 0; i < 3; i++ {
		siteID, err := app.NewSite(siteNames[i], gofakeit.Address().Address)
		if err != nil {
			return fmt.Errorf("seed site: %w", err)
		}
		siteIDs = append(siteIDs, siteID)
		fmt.Printf("Site created: %s\n", siteID)
	}

	fmt.Println("Seeding resources...")
	type seedResource struct {
		name         string
		resourceType model.ResourceType
		unit         *string
		rateAmount   uint64
		costUnit     model.CostUnit
	}
	resourceDefs := []seedResource{
		{"Cable Drum", model.ResourceTypeMaterial, strPtr("metre"), 45, model.CostUnitPerUnit},
		{"LED Lantern", model.ResourceTypeMaterial, strPtr("each"), 12000, model.CostUnitPerUnit},
		{"Cable Jointing Kit", model.ResourceTypeMaterial, strPtr("each"), 3500, model.CostUnitPerUnit},
		{"Drill", model.ResourceTypeTool, nil, 500, model.CostUnitPerHour},
		{"Cable Tester", model.ResourceTypeTool, nil, 800, model.CostUnitPerHour},
		{"Cherry Picker (MEWP)", model.ResourceTypeMechanical, nil, 4500, model.CostUnitPerHour},
		{"Transit Van", model.ResourceTypeMechanical, nil, 2500, model.CostUnitPerHour},
		{"7.5t Truck", model.ResourceTypeMechanical, nil, 6500, model.CostUnitPerHour},
	}
	resourceIDs := make([]*uuid.UUID, 0)
	for _, res := range resourceDefs {
		resourceID, err := app.NewResource(res.name, res.resourceType, res.unit)
		if err != nil {
			return fmt.Errorf("seed resource: %w", err)
		}
		rate := model.NewMoney("GBP", res.rateAmount)
		_, err = app.NewResourceRate(resourceID, rate, res.costUnit, time.Now())
		if err != nil {
			return fmt.Errorf("seed resource rate: %w", err)
		}
		resourceIDs = append(resourceIDs, resourceID)
		fmt.Printf("Resource created: %s @ %s\n", res.name, rate.Serialize())
	}

	materialIDs := resourceIDs[0:3]
	toolIDs := resourceIDs[3:5]
	mechanicalIDs := resourceIDs[5:8]

	fmt.Println("Seeding jobs...")

	seedJob := func(ref, name string, siteID *uuid.UUID, status model.JobStatus, daysAgo int) (*uuid.UUID, error) {
		s := status
		startTime := time.Now().AddDate(0, 0, -daysAgo)
		return app.NewJob(
			ref, name, siteID, qsID, &s, startTime,
			&model.JobOperativeRequirement{ExpectedHeadcount: uint32(gofakeit.IntRange(2, 4))},
			nil,
		)
	}

	seedSession := func(jobID *uuid.UUID, daysAgo int, durationHours float64, submitted bool) (*uuid.UUID, error) {
		start := time.Now().AddDate(0, 0, -daysAgo)
		notes := gofakeit.Sentence(6)
		sessionID, err := app.NewJobSession(jobID, start, &notes)
		if err != nil {
			return nil, err
		}
		if submitted {
			end := start.Add(time.Duration(durationHours * float64(time.Hour)))
			app.UpdateJobSession(sessionID, &end, &end, operativeIDs[0])
		}
		return sessionID, nil
	}

	seedOperativeActual := func(jobID, sessionID, operativeID *uuid.UUID, daysAgo int, hours float64) error {
		arrival := time.Now().AddDate(0, 0, -daysAgo)
		departure := arrival.Add(time.Duration(hours * float64(time.Hour)))
		rate, err := app.GetOperativeRate(operativeID)
		if err != nil {
			return err
		}
		opActualID, err := app.NewJobOperative(jobID, sessionID, operativeID, arrival, rate.RatePerHour)
		if err != nil {
			return err
		}
		cost := rate.RatePerHour.Multiply(hours)
		_, err = app.UpdateJobOperative(opActualID, &departure, &cost)
		return err
	}

	seedResourceActual := func(jobID, sessionID, resourceID *uuid.UUID, qty *float64, duration *float64, arrivalDaysAgo *int) error {
		rate, err := app.GetResourceRate(resourceID)
		if err != nil {
			return err
		}
		var arrivalTime *time.Time
		if arrivalDaysAgo != nil {
			t := time.Now().AddDate(0, 0, -*arrivalDaysAgo)
			arrivalTime = &t
		}
		var cost model.Money
		if qty != nil {
			cost = rate.Rate.Multiply(*qty)
		} else if duration != nil {
			cost = rate.Rate.Multiply(*duration)
		}
		_, err = app.NewJobResource(jobID, sessionID, resourceID, qty, duration, arrivalTime, rate.Rate, &cost)
		return err
	}

	fPtr := func(f float64) *float64 { return &f }
	iPtr := func(i int) *int { return &i }

	// Job 1 — completed, fully costed, approved
	job1ID, err := seedJob("MCN-2025-0091", "Streetlight Column Replacement", siteIDs[0], model.JobStatusApproved, 30)
	if err != nil {
		return fmt.Errorf("seed job1: %w", err)
	}
	session1ID, err := seedSession(job1ID, 30, 8, true)
	if err != nil {
		return fmt.Errorf("seed session1: %w", err)
	}
	seedOperativeActual(job1ID, session1ID, operativeIDs[0], 30, 8)
	seedOperativeActual(job1ID, session1ID, operativeIDs[1], 30, 8)
	seedResourceActual(job1ID, session1ID, materialIDs[1], fPtr(2), nil, nil)
	seedResourceActual(job1ID, session1ID, mechanicalIDs[0], nil, fPtr(8), iPtr(30))
	session1bID, err := seedSession(job1ID, 28, 6, true)
	if err != nil {
		return fmt.Errorf("seed session1b: %w", err)
	}
	seedOperativeActual(job1ID, session1bID, operativeIDs[0], 28, 6)
	seedResourceActual(job1ID, session1bID, materialIDs[0], fPtr(20), nil, nil)
	seedResourceActual(job1ID, session1bID, toolIDs[0], nil, fPtr(6), nil)
	fmt.Printf("Job created (approved): %s\n", job1ID)

	// Job 2 — completed, awaiting approval
	job2ID, err := seedJob("MCN-2026-0012", "EV Charger Cabling", siteIDs[1], model.JobStatusCompleted, 14)
	if err != nil {
		return fmt.Errorf("seed job2: %w", err)
	}
	session2ID, err := seedSession(job2ID, 14, 7, true)
	if err != nil {
		return fmt.Errorf("seed session2: %w", err)
	}
	seedOperativeActual(job2ID, session2ID, operativeIDs[2], 14, 7)
	seedOperativeActual(job2ID, session2ID, operativeIDs[3], 14, 7)
	seedResourceActual(job2ID, session2ID, materialIDs[2], fPtr(5), nil, nil)
	seedResourceActual(job2ID, session2ID, mechanicalIDs[1], nil, fPtr(7), iPtr(14))
	fmt.Printf("Job created (completed): %s\n", job2ID)

	// Job 3 — active, one submitted session, one open session
	job3ID, err := seedJob("MCN-2026-0028", "Road Hole Repair", siteIDs[2], model.JobStatusActive, 10)
	if err != nil {
		return fmt.Errorf("seed job3: %w", err)
	}
	session3aID, err := seedSession(job3ID, 10, 5, true)
	if err != nil {
		return fmt.Errorf("seed session3a: %w", err)
	}
	seedOperativeActual(job3ID, session3aID, operativeIDs[0], 10, 5)
	seedOperativeActual(job3ID, session3aID, operativeIDs[4], 10, 5)
	seedResourceActual(job3ID, session3aID, toolIDs[1], nil, fPtr(5), nil)
	seedResourceActual(job3ID, session3aID, mechanicalIDs[2], nil, fPtr(5), iPtr(10))
	session3bID, err := seedSession(job3ID, 5, 0, false)
	if err != nil {
		return fmt.Errorf("seed session3b: %w", err)
	}
	seedOperativeActual(job3ID, session3bID, operativeIDs[0], 5, 0)
	fmt.Printf("Job created (active, open session): %s\n", job3ID)

	// Job 4 — active, no sessions yet
	job4ID, err := seedJob("MCN-2026-0033", "Traffic Signal Upgrade", siteIDs[0], model.JobStatusActive, 5)
	if err != nil {
		return fmt.Errorf("seed job4: %w", err)
	}
	fmt.Printf("Job created (active, no sessions): %s\n", job4ID)

	// Job 5 — active, no sessions yet
	job5ID, err := seedJob("MCN-2026-0038", "Streetlight Repair", siteIDs[1], model.JobStatusActive, 3)
	if err != nil {
		return fmt.Errorf("seed job5: %w", err)
	}
	fmt.Printf("Job created (active, no sessions): %s\n", job5ID)

	// Job 6 — the recent job you can add stuff to
	job6ID, err := seedJob("MCN-2026-0042", "Streetlight Repair", siteIDs[0], model.JobStatusActive, 1)
	if err != nil {
		return fmt.Errorf("seed job6: %w", err)
	}
	session6ID, err := seedSession(job6ID, 1, 0, false)
	if err != nil {
		return fmt.Errorf("seed session6: %w", err)
	}
	seedOperativeActual(job6ID, session6ID, operativeIDs[0], 1, 0)
	fmt.Printf("Job created (active, open session — add to this one): %s\n", job6ID)

	fmt.Println("Seed complete.")
	fmt.Println("Login: qs@backpack.dev / password123")
	fmt.Println("Recent job ref: MCN-2026-0042")
	return nil
}

func strPtr(s string) *string {
	return &s
}
