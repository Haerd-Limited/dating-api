# How to contribute to this repository 

This document describes Haerd's standard way of developing and contributing to this repository. This is written so that everyone creates features in a unified and consistent way. It describes the standard steps involved when creating a new feature from scratch that doesn't utilise the existing services and also the steps involved when you're creating a new feature that does utilise them. It also describes things such as how to make database changes, 

Notes:
- '//...' represents omitted code to make the examples easier to understand

## Table of Contents
* [Architecture Overview](#architecture-overview)
* [Creating a brand new feature without utilising existing services](#creating-a-brand-new-feature-without-utilising-existing-services)
* [Creating a brand new feature utilising existing services](#creating-a-brand-new-feature-utilising-existing-services)
* [FAQ](#faq)

## Architecture Overview

![clean-arch](../images/clean-arch.png)

In this repository we follow a clean/domain driven design architecture. Mentally you can think of the API/repository being split into 3 layers:
- **The Handler Layer** - Contains handler methods that handles requests sent to the router found in 'internal/http/router/router.go'. Responsible for doing things like :

1. Deserialising JSON into Go request structs and mapping these request structs to the relevant domain struct object which is then passed as an argument to service method calls. 

```golang
//code from 'internal/api/interaction/handler.go'.

func (h *handler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
        //...
		var req dto.SwipesRequest

		// Validates and deserialises the JSON request
		err := request.DecodeAndValidate(r.Body, &req)
        //...

        //Notice how we map requests to domain objects and then pass the domain object to the service method
		result, err := h.interactionService.CreateSwipe(ctx, mapper.SwipesRequestToDomain(req, userID))
        //...
    }
}
```
2. Mapping service method call results into response objects which are included in the response body

```golang
func (h *handler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
        //...
		result, err := h.interactionService.CreateSwipe(ctx, mapper.SwipesRequestToDomain(req, userID))
		if err != nil {
			h.handleServiceErrorResponse(w, r, "CreateSwipe", err)
			return
		}

		render.Json(w, http.StatusCreated, mapper.MapToSwipesResponse(result))
	}
}

```

3. Mapping internal layer(service & repository) errors to the correct status code and user friendly error message

```golang
//code from 'internal/api/interaction/handler.go'.

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, convostorage.ErrClientMsgIDNotUnique):
		return http.StatusBadRequest, "Client message ID must be unique"
	case errors.Is(err, interaction.ErrPromptIDRequiredToLikeUser):
		return http.StatusBadRequest, "Prompt ID is required to like a profile"
	case errors.Is(err, interaction.ErrSelfLike):
		return http.StatusBadRequest, "As much as we promote self love, you cannot like yourself"
	case errors.Is(err, interaction.ErrInvalidAction):
		return http.StatusBadRequest, fmt.Sprintf("Invalid action. Action must be '%s','%s' or '%s'", constants.ActionLike, constants.ActionPass, constants.ActionSuperlike)
	case errors.Is(err, interaction.ErrMissingRequiredFieldsForLikeWithMessage):
		return http.StatusBadRequest, "Sending a like with a message also requires message_type, message,prompt_id and a generated client_msg_id"
	case errors.Is(err, interaction.ErrLikedAVhwUser):
		return http.StatusBadRequest, "You can only superlike or pass a Voices Worth Hearing user"
	case errors.Is(err, interaction.ErrWeeklySuperlikeLimitReached):
		return http.StatusForbidden, "You've already used your superlike this week"
	case errors.Is(err, interaction.ErrInvalidDirection):
		return http.StatusBadRequest, "Invalid direction. Direction must be 'incoming'"
	case errors.Is(err, storage.ErrUserDoesNotExists):
		return http.StatusNotFound, "User does not exist"
	case errors.Is(err, commonErrors.ErrInvalidDOBFormat):
		return http.StatusBadRequest, messages.InvalidDobMsg
	case errors.Is(err, storage2.ErrAlreadySwiped):
		return http.StatusConflict, "You've already swiped on this user"
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg
	}
}

func (h *handler) handleServiceErrorResponse(w http.ResponseWriter, r *http.Request, handlerName string, err error) {
	...
	statusCode, errMsg := mapErrorsToStatusCodeAndUserFriendlyMessages(err)
	...
	render.Json(w, statusCode, commonMappers.ToSimpleErrorResponse(errMsg))
}
```

- **The Service Layer** - Sits in between the **handler layer** and **repository layer**. It's where business logic is applied to domain objects and these domain objects are converted to and from entity objects. Services can also carry out more complex request validation for when "github.com/go-playground/validator/v10" is not enough.

```golang
//code from 'internal/interaction/service.go'

//Notice how the service method parameter is a domain object 'domain.Swipe'
func (is *service) CreateSwipe(ctx context.Context, swipe domain.Swipe) (string, error) {
	//...
	matchable, err := is.interactionRepo.CheckIfMatchable(ctx, swipe.UserID, swipe.TargetUserID)
	if err != nil {
		return "", fmt.Errorf("check if matchable userID=%s targetUserID=%s: %w", swipe.UserID, swipe.TargetUserID, err)
	}

	//...
    // Notice how a function has been made to further validate the domain object incase the handler layer couldn't do all checks.
	err = is.validateSwipe(ctx, swipe, matchable)
	if err != nil {
		return "", fmt.Errorf("validate swipe userID=%s : %w", swipe.UserID, err)
	}

	switch swipe.Action {
	case constants.ActionLike, constants.ActionSuperlike:
		if swipe.Action == constants.ActionSuperlike {
			//...
		}

		if !matchable {
			//...
            //Notice how the swipe domain object is converted/mapped into an entity object before being passed to repository methods.
            err = is.interactionRepo.InsertSwipe(ctx, mapper.SwipeToEntity(swipe), tx.Raw())
			if err != nil {
				return "", fmt.Errorf("insert swipe userID=%s : %w", swipe.UserID, err)
			}

			return ResultSent, nil
		}
		//...

		targetUserSentMeALikeWithAMessage := targetUserSwipe.Message.Valid && targetUserSwipe.MessageType.Valid && targetUserSwipe.IdempotencyKey.Valid
		if targetUserSentMeALikeWithAMessage {
			//...
		}

		userRepliedToLiked := swipe.Message != nil
		if userRepliedToLiked {
			//...
		}

		//...

		return ResultMatched, nil

	case constants.ActionPass:
		//...
		return ResultPassed, nil
	}

	return "", fmt.Errorf("%w: %s", ErrInvalidAction, swipe.Action)
}
```

- **The Repository Layer** - The closest layer to the database. Contains methods that are communicate with the database. Responsible for providing entity objects to the *service layer* so business logic can be applied to it. We generate the entity objects using 'github.com/aarondl/sqlboiler/v4' and the generate files and objects can be found in the 'internal/entity' folder.

```golang
//Code from 'internal/interaction/storage/repository.go'

//Notice how this get method takes parameters and returns the sqlboiler generated entity object.
func (is *repository) GetSwipeByActorIDAndTargetID(ctx context.Context, actorID, targetID string) (*entity.Swipe, error) {
	s, err := entity.Swipes(
		entity.SwipeWhere.ActorID.EQ(actorID),
		entity.SwipeWhere.TargetID.EQ(targetID),
	).One(ctx, is.db)
	if err != nil {
		return nil, err
	}

	return s, nil
}

// Notice how this insert method takes a sqlboiler generated entity object and calls it's Insert method.
func (is *repository) InsertSwipe(ctx context.Context, swipe entity.Swipe, tx *sql.Tx) error {
	err := swipe.Insert(ctx, tx, boil.Infer())
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			switch pqErr.Constraint {
			case "swipes_actor_target_uniq":
				return ErrAlreadySwiped
			}

			return err
		}

		return err
	}

	return nil
}
```

Remember that handlers depend on and import services...  

```golang
//code from 'internal/api/interaction/handler.go'

type handler struct {
	logger             *zap.Logger
	interactionService interaction.Service
}

func NewInteractionHandler(
	logger *zap.Logger,
	interactionService interaction.Service,
) Handler {
	return &handler{
		logger:             logger,
		interactionService: interactionService,
	}
}
```

...Services depend on and import other services and repositories...
```golang
//code from 'internal/interaction/service.go'

type service struct {
	logger              *zap.Logger
	profileService      profile.Service
	conversationService conversation.Service
	interactionRepo     storage.InteractionRepository
	discoverService     discover.Service
	safetyService       safety.Service
	uow                 uow.UoW
	hub                 realtime.Broadcaster
	notificationService notification.Service
}

func NewInteractionService(
	logger *zap.Logger,
	profileService profile.Service,
	conversationService conversation.Service,
	interactionRepo storage.InteractionRepository,
	discoverService discover.Service,
	safetyService safety.Service,
	uow uow.UoW,
	hub realtime.Broadcaster,
	notificationService notification.Service,
) Service {
	return &service{
		logger:              logger,
		interactionRepo:     interactionRepo,
		profileService:      profileService,
		conversationService: conversationService,
		discoverService:     discoverService,
		safetyService:       safetyService,
		uow:                 uow,
		hub:                 hub,
		notificationService: notificationService,
	}
}
```

...and repositories depend on a database. 

```golang
//code from 'internal/interaction/storage/repository.go'
type repository struct {
	db *sqlx.DB
}

func NewInteractionRepository(db *sqlx.DB) InteractionRepository {
	return &repository{
		db: db,
	}
}
```

All of the above allows us to follow SOLID principles and makes it easier to collaborate with others. 

# Creating a brand new feature without utilising existing services 



# Creating a brand new feature utilising existing services 


# FAQ

## What's the purpose of pkg/commonlibrary

This is the folder that contains code that is used across the whole repository. If you find yourself writing some code(it can be a variable, constant or function) inside one file, but then you later realise you need to write it again in a completely different service or file, then you should probably move it to the common library. One simple example is the 'InternalServerErrorMsg' constant. This was something used across ALL handlers. So rather than having to write that same string over and over again in every handler, it was moved to the common library where it can be imported by every handler. This means that now, if we ever want to change that message to have different wording, all we have to do is change it in one place, the 'pkg/commonlibrary/messages/response.go' file and boom its updated everywhere.

```golang
//Code from 'internal/api/interaction/handler.go'
package interaction

import (
	"errors"
	"net/http"

	"github.com/Haerd-Limited/dating-api/pkg/commonlibrary/messages" // common message imported
)

func mapErrorsToStatusCodeAndUserFriendlyMessages(err error) (int, string) {
	switch {
	case errors.Is(err, convostorage.ErrClientMsgIDNotUnique):
		return http.StatusBadRequest, "Client message ID must be unique"
	//...
	default:
		return http.StatusInternalServerError, messages.InternalServerErrorMsg // common InternalServerErrorMsg used
	}
}
```


## What's the purpose of the Makefile





