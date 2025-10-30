# Discover Feed Filtering

This document describes the new filtering capabilities added to the discover feed API.

## Overview

The discover feed now supports filtering profiles by:
- Age Range
- Distance
- Dating Intentions
- Religion
- Ethnicity

Filters can be combined using AND or OR logic.

## API Endpoints

### 1. Original Discover Feed (No Filters)
```
GET /api/v1/discover?limit=10&offset=0
```

### 2. Filtered Discover Feed
```
POST /api/v1/discover/filters
Content-Type: application/json

{
  "limit": 10,
  "offset": 0,
  "filters": {
    "age_range": {
      "min_age": 25,
      "max_age": 35
    },
    "distance": {
      "max_distance_km": 50
    },
    "dating_intentions": {
      "intention_ids": [1, 2, 3]
    },
    "religions": {
      "religion_ids": [1, 2]
    },
    "ethnicities": {
      "ethnicity_ids": [1, 2, 3]
    },
    "operator": "AND"
  }
}
```

## Filter Parameters

### Age Range Filter
- `min_age` (optional): Minimum age in years
- `max_age` (optional): Maximum age in years

### Distance Filter
- `max_distance_km` (optional): Maximum distance in kilometers from user's location

### Dating Intentions Filter
- `intention_ids` (optional): Array of dating intention IDs to include

### Religion Filter
- `religion_ids` (optional): Array of religion IDs to include

### Ethnicity Filter
- `ethnicity_ids` (optional): Array of ethnicity IDs to include

### Filter Operator
- `operator` (optional): How to combine filters - "AND" (default) or "OR"

## Examples

### Filter by Age Range Only
```json
{
  "limit": 20,
  "offset": 0,
  "filters": {
    "age_range": {
      "min_age": 25,
      "max_age": 35
    }
  }
}
```

### Filter by Multiple Criteria (AND)
```json
{
  "limit": 10,
  "offset": 0,
  "filters": {
    "age_range": {
      "min_age": 25,
      "max_age": 35
    },
    "distance": {
      "max_distance_km": 25
    },
    "dating_intentions": {
      "intention_ids": [1, 2]
    },
    "operator": "AND"
  }
}
```

### Filter by Multiple Criteria (OR)
```json
{
  "limit": 10,
  "offset": 0,
  "filters": {
    "religions": {
      "religion_ids": [1, 2]
    },
    "ethnicities": {
      "ethnicity_ids": [1, 2, 3]
    },
    "operator": "OR"
  }
}
```

## Response Format

The response format remains the same as the original discover feed:

```json
{
  "profiles": [
    {
      "user_id": "string",
      "display_name": "string",
      "age": 25,
      "distance_km": 15,
      "gender": "string",
      "dating_intention": "string",
      "religion": "string",
      "ethnicities": ["string"],
      // ... other profile fields
    }
  ]
}
```

## Implementation Notes

1. **Age Calculation**: Age is calculated from the `birthdate` field in the database
2. **Distance Calculation**: Distance is calculated using the `geo` field (latitude/longitude) in the database
3. **Ethnicity Support**: The system supports multiple ethnicities per user via the `user_ethnicities` join table
4. **Performance**: Filters are applied at the database level for optimal performance
5. **Backwards Compatibility**: The original `/discover` endpoint continues to work without any changes

## Database Schema

The filtering implementation uses the following database tables:
- `user_profiles` - Main profile data including age, location, dating intention, religion
- `user_ethnicities` - Many-to-many relationship for user ethnicities
- `dating_intentions` - Lookup table for dating intention options
- `religions` - Lookup table for religion options
- `ethnicities` - Lookup table for ethnicity options
