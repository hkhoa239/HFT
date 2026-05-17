package handlers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"quantalpha/internal/database"
	"quantalpha/internal/models"
)

type AdminHandler struct {
	db database.Querier
}

func NewAdminHandler(db database.Querier) *AdminHandler {
	return &AdminHandler{db: db}
}

func (h *AdminHandler) DeleteJob(c *gin.Context) {
	jobIDStr := c.Param("job_id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.APIResponse{Success: false, Error: "invalid job ID"})
		return
	}

	if err := h.db.DeleteJob(c.Request.Context(), jobID); err != nil {
		c.JSON(http.StatusInternalServerError, models.APIResponse{Success: false, Error: "failed to delete job"})
		return
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true})
}
func (h *AdminHandler) SeedData(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. Seed Users
	users := []struct {
		username string
		role     models.UserRole
		fullName string
	}{
		{"admin", models.RoleAdmin, "System Administrator"},
		{"qr1", models.RoleQR, "Quant Researcher 1"},
		{"qr2", models.RoleQR, "Quant Researcher 2"},
		{"pm1", models.RolePM, "Portfolio Manager 1"},
		{"ds1", models.RoleDS, "Data Scientist 1"},
	}

	passwordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	
	userMap := make(map[string]uuid.UUID)
	for _, u := range users {
		dbUser, _, err := h.db.GetUserByUsername(ctx, u.username)
		if err != nil {
			// Create user
			created, err := h.db.CreateUser(ctx, u.username, string(passwordHash), u.role, u.fullName)
			if err != nil {
				continue
			}
			userMap[u.username] = created.ID
		} else {
			userMap[u.username] = dbUser.ID
		}
	}

	// 2. Seed Alphas (for QR and Admin)
	qrID := userMap["qr1"]
	adminID := userMap["admin"]
	alphaCodes := []struct {
		name   string
		code   string
		author uuid.UUID
	}{
		{"VN30F_Momentum", "def alpha_signal(t, ask1, bq1, bq2, bq3, aq1, aq2, aq3):\n    obi = (bq1 - aq1) / (bq1 + aq1 + 1e-8)\n    return 1 if obi > 0.2 else 0", qrID},
		{"MeanReversion_V1", "def alpha_signal(t, ask1, bq1, bq2, bq3, aq1, aq2, aq3):\n    spread = ask1[0] - bq1\n    return 1 if spread > 2.0 else 0", qrID},
		{"Admin_Strategy", "def alpha_signal(t, ask1, bq1, bq2, bq3, aq1, aq2, aq3):\n    return 1 # Always buy", adminID},
	}

	alphaMap := make(map[string]uuid.UUID)
	for _, a := range alphaCodes {
		created, err := h.db.CreateAlpha(ctx, a.author, a.name, "Sample seeded alpha", a.code)
		if err != nil {
			continue
		}
		h.db.SubmitAlpha(ctx, created.ID)
		alphaMap[a.name] = created.ID
	}

	// 3. Seed Factors
	dsID := userMap["ds1"]
	factors := []struct {
		name string
		freq string
	}{
		{"OBI_111", "1m"},
		{"Spread_Ratio", "5m"},
		{"Volatility_Index", "1h"},
	}
	for _, f := range factors {
		h.db.CreateFactor(ctx, f.name, "Seeded factor", "/data/sample.csv", f.freq, dsID)
	}

	// 4. Seed Backtest Runs with completed status and metrics
	backtestSeeds := []struct {
		name    string
		alphaID uuid.UUID
		pnl     float64
		sharpe  float64
		winRate float64
	}{
		{"VN30F_Momentum", alphaMap["VN30F_Momentum"], 12.4, 2.5, 0.65},
		{"MeanReversion_V1", alphaMap["MeanReversion_V1"], 8.1, 1.8, 0.58},
		{"Admin_Strategy", alphaMap["Admin_Strategy"], -5.2, -0.4, 0.42},
	}

	for _, bs := range backtestSeeds {
		if bs.alphaID == uuid.Nil {
			continue
		}
		run, err := h.db.CreateBacktestRun(ctx, bs.alphaID, adminID, map[string]interface{}{
			"start": "2024-01-01", "end": "2024-01-02", "capital": 1000000,
		})
		if err != nil {
			continue
		}

		// Build PnL curve with enough data points for correlation
		curve := make([]map[string]interface{}, 0, 20)
		for i := 0; i < 20; i++ {
			// Add deterministic non-linear noise depending on alpha name to get a realistic non-zero correlation matrix
			noise := 0.0
			switch bs.name {
			case "VN30F_Momentum":
				noise = float64(i%4)*0.9 - 1.0
			case "MeanReversion_V1":
				noise = float64((i+2)%5)*0.7 - 1.2
			default:
				noise = float64((i*3)%6)*0.5 - 0.8
			}
			curve = append(curve, map[string]interface{}{
				"t":      i + 1,
				"cumPnL": bs.pnl * float64(i) / 19.0 + noise,
			})
		}

		metrics := map[string]interface{}{
			"sharpe_ratio": bs.sharpe, "win_rate": bs.winRate, "total_pnl": bs.pnl,
			"trade_count": 42, "max_drawdown": 2.1,
			"pnl_curve": curve,
		}
		if err := h.db.UpdateBacktestRunStatus(ctx, run.ID, models.JobStatusCompleted, metrics, ""); err != nil {
			log.Printf("SEED: failed to update backtest %s: %v", bs.name, err)
		}
	}

	// 5. Seed Models
	modelsToSeed := []struct {
		name string
		vers string
		metrics map[string]interface{}
	}{
		{"RandomForest_HFT", "v1.0", map[string]interface{}{
			"accuracy": 0.6483, "f1_score": 0.6310, "acc_std": 0.02,
			"precision": 0.652, "recall": 0.613, "auc": 0.706, "log_loss": 0.584,
			"tp": 3214, "fp": 1892, "fn": 2108, "tn": 2786,
		}},
		{"XGBoost_Fast", "v2.1", map[string]interface{}{
			"accuracy": 0.6381, "f1_score": 0.6194, "acc_std": 0.03,
			"precision": 0.641, "recall": 0.598, "auc": 0.692, "log_loss": 0.601,
			"tp": 2986, "fp": 1672, "fn": 2008, "tn": 3334,
		}},
		{"LSTM_Sequence", "v0.5", map[string]interface{}{
			"accuracy": 0.5884, "f1_score": 0.5672, "acc_std": 0.04,
			"precision": 0.582, "recall": 0.553, "auc": 0.634, "log_loss": 0.662,
			"tp": 2760, "fp": 1980, "fn": 2230, "tn": 3030,
		}},
	}
	for _, m := range modelsToSeed {
		h.db.CreateModel(ctx, m.name, m.vers, dsID, "/models/"+m.name+".pkl", m.metrics)
	}

	c.JSON(http.StatusOK, models.APIResponse{Success: true, Data: "Database seeded successfully"})
}
