// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/absmach/magistrala/pkg/apiutil"
	"github.com/absmach/magistrala/pkg/errors"
	repoerr "github.com/absmach/magistrala/pkg/errors/repository"
	"github.com/absmach/magistrala/pkg/postgres"
	"github.com/absmach/magistrala/pkg/roles"
)

var _ roles.Repository = (*RolesSvcRepo)(nil)

type RolesSvcRepo struct {
	db postgres.Database
}

// NewRolesSvcRepository instantiates a PostgreSQL
// implementation of Roles repository.
func NewRolesSvcRepository(db postgres.Database) RolesSvcRepo {
	return RolesSvcRepo{
		db: db,
	}
}

type dbPage struct {
	ID       string `db:"id"`
	Name     string `db:"name"`
	EntityID string `db:"entity_id"`
	Limit    uint64 `db:"limit"`
	Offset   uint64 `db:"offset"`
}
type dbRole struct {
	ID        string       `db:"id"`
	Name      string       `db:"name"`
	EntityID  string       `db:"entity_id"`
	CreatedBy *string      `db:"created_by"`
	CreatedAt sql.NullTime `db:"created_at"`
	UpdatedBy *string      `db:"updated_by"`
	UpdatedAt sql.NullTime `db:"updated_at"`
}

type dbRoleOperation struct {
	RoleID    string `db:"role_id"`
	Operation string `db:"operation"`
}

type dbRoleMember struct {
	RoleID string `db:"role_id"`
	Member string `db:"member"`
}

func toDBRoles(role roles.Role) dbRole {
	var createdBy *string
	if role.CreatedBy != "" {
		createdBy = &role.UpdatedBy
	}
	var createdAt sql.NullTime
	if role.CreatedAt != (time.Time{}) && !role.CreatedAt.IsZero() {
		createdAt = sql.NullTime{Time: role.CreatedAt, Valid: true}
	}

	var updatedBy *string
	if role.UpdatedBy != "" {
		updatedBy = &role.UpdatedBy
	}
	var updatedAt sql.NullTime
	if role.UpdatedAt != (time.Time{}) && !role.UpdatedAt.IsZero() {
		updatedAt = sql.NullTime{Time: role.UpdatedAt, Valid: true}
	}

	return dbRole{
		ID:        role.ID,
		Name:      role.Name,
		EntityID:  role.EntityID,
		CreatedBy: createdBy,
		CreatedAt: createdAt,
		UpdatedBy: updatedBy,
		UpdatedAt: updatedAt,
	}
}

func toRole(r dbRole) roles.Role {

	var createdBy string
	if r.CreatedBy != nil {
		createdBy = *r.CreatedBy
	}
	var createdAt time.Time
	if r.CreatedAt.Valid {
		createdAt = r.CreatedAt.Time
	}

	var updatedBy string
	if r.UpdatedBy != nil {
		updatedBy = *r.UpdatedBy
	}
	var updatedAt time.Time
	if r.UpdatedAt.Valid {
		updatedAt = r.UpdatedAt.Time
	}

	return roles.Role{
		ID:        r.ID,
		Name:      r.Name,
		EntityID:  r.EntityID,
		CreatedBy: createdBy,
		CreatedAt: createdAt,
		UpdatedBy: updatedBy,
		UpdatedAt: updatedAt,
	}

}
func (repo *RolesSvcRepo) AddRoles(ctx context.Context, rps []roles.RoleProvision) ([]roles.Role, error) {

	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return []roles.Role{}, errors.Wrap(repoerr.ErrCreateEntity, err)
	}
	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Wrap(errors.Wrap(apiutil.ErrRollbackTx, errRollback), err)
			}
		}
	}()

	var retRoles []roles.Role

	for _, rp := range rps {

		q := `INSERT INTO roles (id, name, entity_id, created_by, created_at, updated_by, updated_at)
        VALUES (:id :name :entity_id :created_by :created_at :updated_by :updated_at)
        RETURNING id, name, entity_id, created_by, created_at, updated_by, updated_at`

		row, err := tx.NamedQuery(q, toDBRoles(rp.Role))
		if err != nil {
			return []roles.Role{}, postgres.HandleError(repoerr.ErrCreateEntity, err)
		}
		defer row.Close()
		row.Next()
		dbr := dbRole{}
		if err := row.StructScan(&dbr); err != nil {
			return []roles.Role{}, errors.Wrap(repoerr.ErrFailedOpDB, err)
		}
		retRoles = append(retRoles, toRole(dbr))

		if len(rp.OptionalOperations) > 0 {
			opq := `INSERT INTO role_operations (role_id, operation)
        				VALUES (:role_id, :operation)
        				RETURNING role_id, operation`

			rOps := []dbRoleOperation{}
			for _, op := range rp.OptionalOperations {
				rOps = append(rOps, dbRoleOperation{
					RoleID:    rp.ID,
					Operation: string(op),
				})
			}
			if _, err := tx.NamedExecContext(ctx, opq, rOps); err != nil {
				return []roles.Role{}, postgres.HandleError(repoerr.ErrCreateEntity, err)
			}
		}

		if len(rp.OptionalMembers) > 0 {
			mq := `INSERT INTO role_members (role_id, member)
					VALUES (:role_id, :member)
					RETURNING role_id, member`

			rMems := []dbRoleMember{}
			for _, m := range rp.OptionalMembers {
				rMems = append(rMems, dbRoleMember{
					RoleID: rp.ID,
					Member: m,
				})
			}
			if _, err := tx.NamedExecContext(ctx, mq, rMems); err != nil {
				return []roles.Role{}, postgres.HandleError(repoerr.ErrCreateEntity, err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return []roles.Role{}, postgres.HandleError(repoerr.ErrCreateEntity, err)
	}

	return retRoles, nil
}

func (repo *RolesSvcRepo) RemoveRoles(ctx context.Context, roleIDs []string) error {
	q := "DELETE FROM roles  WHERE id IN (:role_id) ;"

	params := map[string]interface{}{
		"role_id": roleIDs,
	}
	result, err := repo.db.ExecContext(ctx, q, params)
	if err != nil {
		return postgres.HandleError(repoerr.ErrRemoveEntity, err)
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		return repoerr.ErrNotFound
	}

	return nil
}

// Update only role name, don't update ID
func (repo *RolesSvcRepo) UpdateRole(ctx context.Context, role roles.Role) (roles.Role, error) {
	var query []string
	var upq string
	if role.Name != "" {
		query = append(query, "name = :name,")
	}

	if len(query) > 0 {
		upq = strings.Join(query, " ")
	}

	q := fmt.Sprintf(`UPDATE roles SET %s updated_at = :updated_at, updated_by = :updated_by
        WHERE id = :id
        RETURNING id, name, entity_id, created_by, created_at, updated_by, updated_at`,
		upq)

	row, err := repo.db.NamedQueryContext(ctx, q, toDBRoles(role))

	if err != nil {
		return roles.Role{}, postgres.HandleError(repoerr.ErrUpdateEntity, err)
	}
	defer row.Close()

	dbr := dbRole{}
	if row.Next() {
		if err := row.StructScan(&dbr); err != nil {
			return roles.Role{}, errors.Wrap(repoerr.ErrUpdateEntity, err)
		}
		return toRole(dbr), nil
	}

	return roles.Role{}, repoerr.ErrNotFound
}

func (repo *RolesSvcRepo) RetrieveRole(ctx context.Context, roleID string) (roles.Role, error) {
	q := `SELECT id, name, entity_id, created_by, created_at, updated_by, updated_at
        FROM roles WHERE id = :id`

	dbr := dbRole{
		ID: roleID,
	}

	rows, err := repo.db.NamedQueryContext(ctx, q, dbr)
	if err != nil {
		return roles.Role{}, postgres.HandleError(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	dbr = dbRole{}
	if rows.Next() {
		if err = rows.StructScan(&dbr); err != nil {
			return roles.Role{}, postgres.HandleError(repoerr.ErrViewEntity, err)
		}

		return toRole(dbr), nil
	}

	return roles.Role{}, repoerr.ErrNotFound
}

func (repo *RolesSvcRepo) RetrieveRoleByEntityIDAndName(ctx context.Context, entityID, roleName string) (roles.Role, error) {
	q := `SELECT id, name, entity_id, created_by, created_at, updated_by, updated_at
        FROM roles WHERE entity_id = :entity_id and name = :name`

	dbr := dbRole{
		EntityID: entityID,
		Name:     roleName,
	}

	rows, err := repo.db.NamedQueryContext(ctx, q, dbr)
	if err != nil {
		return roles.Role{}, postgres.HandleError(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	dbr = dbRole{}
	if rows.Next() {
		if err = rows.StructScan(&dbr); err != nil {
			return roles.Role{}, postgres.HandleError(repoerr.ErrViewEntity, err)
		}

		return toRole(dbr), nil
	}

	return roles.Role{}, repoerr.ErrNotFound
}
func (repo *RolesSvcRepo) RetrieveAllRoles(ctx context.Context, entityID string, limit, offset uint64) (roles.RolePage, error) {
	q := `SELECT id, name, entity_id, created_by, created_at, updated_by, updated_at
        FROM roles WHERE entity_id = :entity_id ORDER BY created_at LIMIT :limit OFFSET :offset;`

	dbp := dbPage{
		EntityID: entityID,
		Limit:    limit,
		Offset:   offset,
	}

	rows, err := repo.db.NamedQueryContext(ctx, q, dbp)
	if err != nil {
		return roles.RolePage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	var items []roles.Role
	for rows.Next() {
		dbr := dbRole{}
		if err := rows.StructScan(&dbr); err != nil {
			return roles.RolePage{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}

		items = append(items, toRole(dbr))
	}
	cq := `SELECT COUNT(*) FROM roles WHERE entity_id = :entity_id ORDER BY created_at LIMIT :limit OFFSET :offset;`

	total, err := postgres.Total(ctx, repo.db, cq, dbp)
	if err != nil {
		return roles.RolePage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	page := roles.RolePage{
		Roles:  items,
		Total:  total,
		Offset: offset,
		Limit:  limit,
	}

	return page, nil
}

func (repo *RolesSvcRepo) RoleAddOperation(ctx context.Context, role roles.Role, operations []roles.Operation) (ops []roles.Operation, err error) {

	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return []roles.Operation{}, errors.Wrap(repoerr.ErrCreateEntity, err)
	}
	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Wrap(errors.Wrap(apiutil.ErrRollbackTx, errRollback), err)
			}
		}
	}()

	opq := `INSERT INTO role_operations (role_id, operation)
	VALUES (:role_id, :operation)
	RETURNING role_id, operation`

	rOps := []dbRoleOperation{}
	for _, op := range operations {
		rOps = append(rOps, dbRoleOperation{
			RoleID:    role.ID,
			Operation: string(op),
		})
	}
	if _, err := tx.NamedExecContext(ctx, opq, rOps); err != nil {
		return []roles.Operation{}, postgres.HandleError(repoerr.ErrCreateEntity, err)
	}

	upq := `UPDATE roles SET updated_at = :updated_at, updated_by = :updated_by WHERE id = :id;`
	if _, err := tx.NamedExecContext(ctx, upq, toDBRoles(role)); err != nil {
		return []roles.Operation{}, postgres.HandleError(repoerr.ErrCreateEntity, err)
	}

	if err := tx.Commit(); err != nil {
		return []roles.Operation{}, postgres.HandleError(repoerr.ErrCreateEntity, err)
	}

	return repo.RoleListOperations(ctx, role.ID)
}

func (repo *RolesSvcRepo) RoleListOperations(ctx context.Context, roleID string) ([]roles.Operation, error) {
	q := `SELECT role_id, operation FROM role_operations WHERE role_id = :role_id ;`

	dbrop := dbRoleOperation{
		RoleID: roleID,
	}

	rows, err := repo.db.NamedQueryContext(ctx, q, dbrop)
	if err != nil {
		return []roles.Operation{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	var items []roles.Operation
	for rows.Next() {
		dbrop = dbRoleOperation{}
		if err := rows.StructScan(&dbrop); err != nil {
			return []roles.Operation{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}

		items = append(items, roles.Operation(dbrop.Operation))
	}
	return items, nil
}

func (repo *RolesSvcRepo) RoleCheckOperationsExists(ctx context.Context, roleID string, operations []roles.Operation) (bool, error) {
	q := ` SELECT COUNT(*) FROM role_operations WHERE role_id = :role_id AND operation IN (:operations)`

	params := map[string]interface{}{
		"role_id":    roleID,
		"operations": operations,
	}
	var count int
	query, err := repo.db.NamedQueryContext(ctx, q, params)
	if err != nil {
		return false, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	defer query.Close()

	if query.Next() {
		if err := query.Scan(&count); err != nil {
			return false, errors.Wrap(repoerr.ErrViewEntity, err)
		}
	}

	// Check if the count matches the number of operations provided
	if count != len(operations) {
		return false, nil
	}

	return true, nil
}

func (repo *RolesSvcRepo) RoleRemoveOperations(ctx context.Context, role roles.Role, operations []roles.Operation) (err error) {

	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}
	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Wrap(errors.Wrap(apiutil.ErrRollbackTx, errRollback), err)
			}
		}
	}()

	q := `DELETE FROM role_operations WHERE role_id = :role_id AND operation IN (:operations)`

	params := map[string]interface{}{
		"role_id":    role.ID,
		"operations": operations,
	}

	if _, err := tx.NamedExecContext(ctx, q, params); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	upq := `UPDATE roles SET updated_at = :updated_at, updated_by = :updated_by WHERE id = :id;`
	if _, err := tx.NamedExecContext(ctx, upq, toDBRoles(role)); err != nil {
		return postgres.HandleError(repoerr.ErrRemoveEntity, err)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	return nil
}

func (repo *RolesSvcRepo) RoleRemoveAllOperations(ctx context.Context, role roles.Role) error {
	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}
	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Wrap(errors.Wrap(apiutil.ErrRollbackTx, errRollback), err)
			}
		}
	}()

	q := `DELETE FROM role_operations WHERE role_id = :role_id `

	dbrop := dbRoleOperation{RoleID: role.ID}

	if _, err := tx.NamedExecContext(ctx, q, dbrop); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	upq := `UPDATE roles SET updated_at = :updated_at, updated_by = :updated_by WHERE id = :id;`
	if _, err := tx.NamedExecContext(ctx, upq, toDBRoles(role)); err != nil {
		return postgres.HandleError(repoerr.ErrRemoveEntity, err)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	return nil
}

func (repo *RolesSvcRepo) RoleAddMembers(ctx context.Context, role roles.Role, members []string) ([]string, error) {
	mq := `INSERT INTO role_members (role_id, member)
        VALUES (:role_id, :member)
        RETURNING role_id, member`

	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return []string{}, errors.Wrap(repoerr.ErrCreateEntity, err)
	}
	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Wrap(errors.Wrap(apiutil.ErrRollbackTx, errRollback), err)
			}
		}
	}()

	rMems := []dbRoleMember{}
	for _, m := range members {
		rMems = append(rMems, dbRoleMember{
			RoleID: role.ID,
			Member: m,
		})
	}
	if _, err := tx.NamedExecContext(ctx, mq, rMems); err != nil {
		return []string{}, postgres.HandleError(repoerr.ErrCreateEntity, err)
	}

	upq := `UPDATE roles SET updated_at = :updated_at, updated_by = :updated_by WHERE id = :id;`
	if _, err := tx.NamedExecContext(ctx, upq, toDBRoles(role)); err != nil {
		return []string{}, postgres.HandleError(repoerr.ErrCreateEntity, err)
	}

	if err := tx.Commit(); err != nil {
		return []string{}, postgres.HandleError(repoerr.ErrCreateEntity, err)
	}

	return members, nil
}

func (repo *RolesSvcRepo) RoleListMembers(ctx context.Context, roleID string, limit, offset uint64) (roles.MembersPage, error) {
	q := `SELECT role_id, member FROM role_members WHERE role_id = :role_id ORDER BY created_at LIMIT :limit OFFSET :offset;`

	dbrmems := dbRoleMember{
		RoleID: roleID,
	}

	rows, err := repo.db.NamedQueryContext(ctx, q, dbrmems)
	if err != nil {
		return roles.MembersPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		dbrmems = dbRoleMember{}
		if err := rows.StructScan(&dbrmems); err != nil {
			return roles.MembersPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
		}

		items = append(items, dbrmems.Member)
	}

	cq := `SELECT COUNT(*) FROM role_members WHERE role_id = :role_id ORDER BY created_at LIMIT :limit OFFSET :offset;`

	total, err := postgres.Total(ctx, repo.db, cq, dbrmems)
	if err != nil {
		return roles.MembersPage{}, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	return roles.MembersPage{
		Members: items,
		Total:   total,
		Offset:  offset,
		Limit:   limit,
	}, nil

}

func (repo *RolesSvcRepo) RoleCheckMembersExists(ctx context.Context, roleID string, members []string) (bool, error) {
	q := ` SELECT COUNT(*) FROM role_members WHERE role_id = :role_id AND operation IN (:members)`

	params := map[string]interface{}{
		"role_id": roleID,
		"members": members,
	}
	var count int
	query, err := repo.db.NamedQueryContext(ctx, q, params)
	if err != nil {
		return false, errors.Wrap(repoerr.ErrViewEntity, err)
	}

	defer query.Close()

	if query.Next() {
		if err := query.Scan(&count); err != nil {
			return false, errors.Wrap(repoerr.ErrViewEntity, err)
		}
	}

	// Check if the count matches the number of operations provided
	if count != len(members) {
		return false, nil
	}

	return true, nil
}

func (repo *RolesSvcRepo) RoleRemoveMembers(ctx context.Context, role roles.Role, members []string) (err error) {
	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}
	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Wrap(errors.Wrap(apiutil.ErrRollbackTx, errRollback), err)
			}
		}
	}()

	q := `DELETE FROM role_members WHERE role_id = :role_id AND member_id IN (:member_id)`

	params := map[string]interface{}{
		"role_id":   role.ID,
		"member_id": members,
	}

	if _, err := tx.NamedExecContext(ctx, q, params); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	upq := `UPDATE roles SET updated_at = :updated_at, updated_by = :updated_by WHERE id = :id;`
	if _, err := tx.NamedExecContext(ctx, upq, toDBRoles(role)); err != nil {
		return postgres.HandleError(repoerr.ErrRemoveEntity, err)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}
	return nil
}

func (repo *RolesSvcRepo) RoleRemoveAllMembers(ctx context.Context, role roles.Role) (err error) {
	tx, err := repo.db.BeginTxx(ctx, nil)
	if err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}
	defer func() {
		if err != nil {
			if errRollback := tx.Rollback(); errRollback != nil {
				err = errors.Wrap(errors.Wrap(apiutil.ErrRollbackTx, errRollback), err)
			}
		}
	}()
	q := `DELETE FROM role_members WHERE role_id = :role_id `

	dbrop := dbRoleOperation{RoleID: role.ID}

	if _, err := repo.db.NamedExecContext(ctx, q, dbrop); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}

	upq := `UPDATE roles SET updated_at = :updated_at, updated_by = :updated_by WHERE id = :id;`
	if _, err := tx.NamedExecContext(ctx, upq, toDBRoles(role)); err != nil {
		return postgres.HandleError(repoerr.ErrRemoveEntity, err)
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(repoerr.ErrRemoveEntity, err)
	}
	return nil
}
