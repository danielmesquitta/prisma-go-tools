package tables

type Table string

const (
	TableTransaction    Table = "transactions"
	TableInvestment     Table = "investments"
	TableAccount        Table = "accounts"
	TableCreditCard     Table = "credit_cards"
	TableInstitution    Table = "institutions"
	TableBudget         Table = "budgets"
	TableUser           Table = "users"
	TableBudgetCategory Table = "budget_categories"
	TableCategory       Table = "categories"
)
