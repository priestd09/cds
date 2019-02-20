package api

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	yaml "gopkg.in/yaml.v2"

	"github.com/ovh/cds/engine/api/action"
	"github.com/ovh/cds/engine/api/event"
	"github.com/ovh/cds/engine/api/group"
	"github.com/ovh/cds/engine/api/project"
	"github.com/ovh/cds/engine/service"
	"github.com/ovh/cds/sdk"
	"github.com/ovh/cds/sdk/exportentities"
)

const (
	contextAction contextKey = iota
)

func (api *API) middlewareAction(needAdmin bool) func(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) (context.Context, error) {
		// try to get action for given path that match user's groups with/without admin grants
		vars := mux.Vars(r)

		groupName := vars["groupName"]
		actionName := vars["actionName"]

		if groupName == "" || actionName == "" {
			return nil, sdk.WrapError(sdk.ErrWrongRequest, "invalid given group or action name")
		}

		u := deprecatedGetUser(ctx)

		// check that group exists
		g, err := group.LoadGroup(api.mustDB(), groupName)
		if err != nil {
			return nil, err
		}

		if needAdmin {
			if err := group.CheckUserIsGroupAdmin(g, u); err != nil {
				return nil, err
			}
		} else {
			if err := group.CheckUserIsGroupMember(g, u); err != nil {
				return nil, err
			}
		}

		a, err := action.GetTypeDefaultByNameAndGroupID(api.mustDB(), actionName, g.ID)
		if err != nil {
			return nil, err
		}
		if a == nil {
			return nil, sdk.WithStack(sdk.ErrNoAction)
		}

		return context.WithValue(ctx, contextAction, a), nil
	}
}

func getAction(c context.Context) *sdk.Action {
	i := c.Value(contextAction)
	if i == nil {
		return nil
	}
	a, ok := i.(*sdk.Action)
	if !ok {
		return nil
	}
	return a
}

func (api *API) getActionsHandler() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		u := deprecatedGetUser(ctx)

		var as []sdk.Action
		var err error
		if u.Admin {
			as, err = action.LoadAllTypeDefault(api.mustDB())
		} else {
			as, err = action.LoadAllTypeDefaultByGroupIDs(api.mustDB(), append(
				sdk.GroupsToIDs(u.Groups), group.SharedInfraGroup.ID))
		}
		if err != nil {
			return err
		}

		return service.WriteJSON(w, as, http.StatusOK)
	}
}

func (api *API) getActionsForProjectHandler() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		vars := mux.Vars(r)
		key := vars[permProjectKey]

		proj, err := project.Load(api.mustDB(), api.Cache, key, deprecatedGetUser(ctx), project.LoadOptions.WithGroups)
		if err != nil {
			return sdk.WrapError(err, "unable to load projet %s", key)
		}

		groupIDs := make([]int64, len(proj.ProjectGroups))
		for i := range proj.ProjectGroups {
			groupIDs[i] = proj.ProjectGroups[i].Group.ID
		}

		as, err := action.LoadAllTypeBuiltInOrPluginOrDefaultForGroupIDs(api.mustDB(),
			append(groupIDs, group.SharedInfraGroup.ID))
		if err != nil {
			return err
		}

		return service.WriteJSON(w, as, http.StatusOK)
	}
}

func (api *API) getActionsForGroupHandler() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		groupID, err := requestVarInt(r, "groupID")
		if err != nil {
			return err
		}

		// check that the group exists and user is part of the group
		g, err := group.LoadGroupByID(api.mustDB(), groupID)
		if err != nil {
			return err
		}

		u := deprecatedGetUser(ctx)

		if err := group.CheckUserIsGroupMember(g, u); err != nil {
			return err
		}

		as, err := action.LoadAllTypeBuiltInOrPluginOrDefaultForGroupIDs(api.mustDB(),
			[]int64{g.ID, group.SharedInfraGroup.ID})
		if err != nil {
			return err
		}

		return service.WriteJSON(w, as, http.StatusOK)
	}
}

func (api *API) postActionHandler() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		var data sdk.Action
		if err := service.UnmarshalBody(r, &data); err != nil {
			return err
		}
		if err := data.IsValid(); err != nil {
			return err
		}

		// check that the group exists and user is admin for group id
		grp, err := group.LoadGroupByID(api.mustDB(), *data.GroupID)
		if err != nil {
			return err
		}

		u := deprecatedGetUser(ctx)

		if err := group.CheckUserIsGroupAdmin(grp, u); err != nil {
			return err
		}

		tx, err := api.mustDB().Begin()
		if err != nil {
			return sdk.WithStack(err)
		}
		defer tx.Rollback()

		// check that no action already exists for same group/name
		current, err := action.GetTypeDefaultByNameAndGroupID(tx, data.Name, grp.ID)
		if err != nil {
			return err
		}
		if current != nil {
			return sdk.NewErrorFrom(sdk.ErrAlreadyExist, "an action already exists for given name on this group")
		}

		// only default action can be posted or updated
		data.Type = sdk.DefaultAction
		data.Enabled = true

		// check that given children exists and can be used
		if err := action.CheckChildrenForGroupIDs(tx, &data, []int64{group.SharedInfraGroup.ID, grp.ID}); err != nil {
			return err
		}

		// inserts action and components
		if err := action.Insert(tx, &data); err != nil {
			return err
		}

		if err := tx.Commit(); err != nil {
			return sdk.WithStack(err)
		}

		new, err := action.LoadByID(api.mustDB(), data.ID)
		if err != nil {
			return err
		}

		event.PublishActionAdd(*new, u)

		new.Editable = true

		return service.WriteJSON(w, new, http.StatusCreated)
	}
}

func (api *API) getActionHandler() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		ctx, err := api.middlewareAction(false)(ctx, w, r)
		if err != nil {
			return err
		}

		a := getAction(ctx)

		if err := action.FullView.Exec(api.mustDB(), a); err != nil {
			return err
		}
		if err := group.CheckUserIsGroupAdmin(a.Group, deprecatedGetUser(ctx)); err == nil {
			a.Editable = true
		}

		return service.WriteJSON(w, a, http.StatusOK)
	}
}

func (api *API) putActionHandler() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		ctx, err := api.middlewareAction(true)(ctx, w, r)
		if err != nil {
			return err
		}

		old := getAction(ctx)

		if err := action.FullView.Exec(api.mustDB(), old); err != nil {
			return err
		}

		var data sdk.Action
		if err := service.UnmarshalBody(r, &data); err != nil {
			return err
		}
		if err := data.IsValid(); err != nil {
			return err
		}

		tx, err := api.mustDB().Begin()
		if err != nil {
			return sdk.WrapError(err, "cannot begin transaction")
		}
		defer tx.Rollback()

		u := deprecatedGetUser(ctx)

		grp, err := group.LoadGroupByID(tx, *data.GroupID)
		if err != nil {
			return err
		}

		if *old.GroupID != *data.GroupID || old.Name != data.Name {
			// check that the group exists and user is admin for group id
			if err := group.CheckUserIsGroupAdmin(grp, u); err != nil {
				return err
			}

			// check that no action already exists for same group/name
			current, err := action.GetTypeDefaultByNameAndGroupID(tx, data.Name, grp.ID)
			if err != nil {
				return err
			}
			if current != nil {
				return sdk.NewErrorFrom(sdk.ErrAlreadyExist, "an action already exists for given name on this group")
			}
		}

		// only default action can be posted or updated
		data.ID = old.ID
		data.Type = sdk.DefaultAction
		data.Enabled = true

		// check that given children exists and can be used, and no loop exists
		if err := action.CheckChildrenForGroupIDsWithLoop(tx, &data, []int64{group.SharedInfraGroup.ID, grp.ID}); err != nil {
			return err
		}

		if err = action.Update(tx, &data); err != nil {
			return sdk.WrapError(err, "cannot update action")
		}

		if err = tx.Commit(); err != nil {
			return sdk.WrapError(err, "cannot commit transaction")
		}

		new, err := action.LoadByID(api.mustDB(), data.ID)
		if err != nil {
			return err
		}

		event.PublishActionUpdate(*old, *new, u)

		new.Editable = true

		return service.WriteJSON(w, new, http.StatusOK)
	}
}

func (api *API) deleteActionHandler() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		ctx, err := api.middlewareAction(true)(ctx, w, r)
		if err != nil {
			return err
		}

		a := getAction(ctx)

		tx, err := api.mustDB().Begin()
		if err != nil {
			return sdk.WrapError(err, "cannot start transaction")
		}
		defer tx.Rollback()

		used, err := action.Used(tx, a.ID)
		if err != nil {
			return err
		}
		if used {
			return sdk.NewErrorFrom(sdk.ErrForbidden, "cannot delete action %s is used in other actions or pipelines", a.Name)
		}

		if err := action.Delete(tx, a); err != nil {
			return sdk.WrapError(err, "cannot delete action %s", a.Name)
		}

		if err := tx.Commit(); err != nil {
			return sdk.WrapError(err, "cannot commit transaction")
		}

		return nil
	}
}

func (api *API) getActionAuditHandler() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		ctx, err := api.middlewareAction(false)(ctx, w, r)
		if err != nil {
			return err
		}

		a := getAction(ctx)

		as, err := action.GetAuditsByActionIDsAndEventTypes(api.mustDB(),
			[]int64{a.ID}, []string{"ActionAdd", "ActionUpdate"})
		if err != nil {
			return err
		}

		// convert all audits to export entities yaml
		for i := range as {
			as[i].DataType = "yaml"
			if as[i].DataBefore != "" {
				var before sdk.Action
				if err := json.Unmarshal([]byte(as[i].DataBefore), &before); err != nil {
					return sdk.WrapError(err, "cannot parse action audit")
				}

				ea := exportentities.NewAction(before)
				buf, err := yaml.Marshal(ea)
				if err != nil {
					return sdk.WrapError(err, "cannot parse action audit")
				}

				as[i].DataBefore = string(buf)
			}

			if as[i].DataAfter != "" {
				var after sdk.Action
				if err := json.Unmarshal([]byte(as[i].DataAfter), &after); err != nil {
					return sdk.WrapError(err, "cannot parse action audit")
				}

				ea := exportentities.NewAction(after)
				buf, err := yaml.Marshal(ea)
				if err != nil {
					return sdk.WrapError(err, "cannot parse action audit")
				}

				as[i].DataAfter = string(buf)
			}
		}

		return service.WriteJSON(w, as, http.StatusOK)
	}
}

func (api *API) getActionUsageHandler() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		ctx, err := api.middlewareAction(false)(ctx, w, r)
		if err != nil {
			return err
		}

		a := getAction(ctx)

		// TODO filter pipelines and action for current user's groups
		pus, err := action.GetPipelineUsages(api.mustDB(), group.SharedInfraGroup.ID, a.ID)
		if err != nil {
			return err
		}
		aus, err := action.GetActionUsages(api.mustDB(), group.SharedInfraGroup.ID, a.ID)
		if err != nil {
			return err
		}

		return service.WriteJSON(w, action.Usage{
			Pipelines: pus,
			Actions:   aus,
		}, http.StatusOK)
	}
}

func (api *API) getActionExportHandler() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		ctx, err := api.middlewareAction(false)(ctx, w, r)
		if err != nil {
			return err
		}

		a := getAction(ctx)

		format := FormString(r, "format")
		if format == "" {
			format = "yaml"
		}

		f, err := exportentities.GetFormat(format)
		if err != nil {
			return err
		}

		if err := action.FullView.Exec(api.mustDB(), a); err != nil {
			return err
		}

		if err := action.Export(*a, f, w); err != nil {
			return err
		}

		w.Header().Add("Content-Type", exportentities.GetContentType(f))
		return nil
	}
}

// importActionHandler insert OR update an existing action.
func (api *API) importActionHandler() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			return err
		}
		defer r.Body.Close()

		contentType := r.Header.Get("Content-Type")
		if contentType == "" {
			contentType = http.DetectContentType(body)
		}

		var ea exportentities.Action
		switch contentType {
		case "application/json":
			err = json.Unmarshal(body, &ea)
		case "application/x-yaml", "text/x-yam":
			err = yaml.Unmarshal(body, &ea)
		default:
			return sdk.NewErrorFrom(sdk.ErrWrongRequest, "unsupported content-type: %s", contentType)
		}
		if err != nil {
			return sdk.NewError(sdk.ErrWrongRequest, err)
		}

		data, err := ea.Action()
		if err != nil {
			return sdk.NewError(sdk.ErrWrongRequest, err)
		}

		tx, err := api.mustDB().Begin()
		if err != nil {
			return sdk.WithStack(err)
		}
		defer tx.Rollback()

		// set group id on given action, if no group given use shared.infra fo backward compatibility
		// current user should be admin if the group
		var grp *sdk.Group
		if data.Group.Name == sdk.SharedInfraGroupName {
			grp = group.SharedInfraGroup
		} else {
			grp, err = group.LoadGroupByName(tx, data.Group.Name)
			if err != nil {
				return err
			}
		}

		u := deprecatedGetUser(ctx)

		if err := group.CheckUserIsGroupAdmin(grp, u); err != nil {
			return err
		}

		data.GroupID = &grp.ID

		// set action id for children based on action name and group name
		// if no group name given for child, first search an action for shared.infra for backward compatibility
		// else search a builtin or plugin action
		for i := range data.Actions {
			a, err := action.RetrieveForGroupAndName(tx, data.Actions[i].Group, data.Actions[i].Name)
			if err != nil {
				return err
			}
			data.Actions[i].ID = a.ID
		}

		// check data validity
		if err := data.IsValid(); err != nil {
			return err
		}

		// check if action exists in database
		old, err := action.LoadTypeDefaultByNameAndGroupID(api.mustDB(), data.Name, grp.ID)
		if err != nil {
			return err
		}
		exists := old != nil

		// update or insert depending action if action exists
		if exists {
			data.ID = old.ID

			// check that given children exists and can be used, and no loop exists
			if err := action.CheckChildrenForGroupIDsWithLoop(tx, &data, []int64{group.SharedInfraGroup.ID, grp.ID}); err != nil {
				return err
			}

			if err = action.Update(tx, &data); err != nil {
				return sdk.WrapError(err, "cannot update action")
			}
		} else {
			// check that given children exists and can be used
			if err := action.CheckChildrenForGroupIDs(tx, &data, []int64{group.SharedInfraGroup.ID, grp.ID}); err != nil {
				return err
			}

			// inserts action and components
			if err := action.Insert(tx, &data); err != nil {
				return err
			}
		}

		if err := tx.Commit(); err != nil {
			return sdk.WithStack(err)
		}

		new, err := action.LoadByID(api.mustDB(), data.ID)
		if err != nil {
			return err
		}

		if exists {
			event.PublishActionUpdate(*old, *new, u)
		} else {
			event.PublishActionAdd(*new, u)
		}

		code := http.StatusCreated
		if exists {
			code = http.StatusOK
		}
		return service.WriteJSON(w, new, code)
	}
}

func (api *API) getActionsRequirements() service.Handler {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
		rs, err := action.GetRequirementsDistinctBinary(api.mustDB())
		if err != nil {
			return sdk.WrapError(err, "cannot load action requirements")
		}

		return service.WriteJSON(w, rs, http.StatusOK)
	}
}
