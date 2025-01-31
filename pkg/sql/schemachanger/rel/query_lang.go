// Copyright 2021 The Cockroach Authors.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package rel

import "reflect"

// Clause is the basic building block of a query. A query is defined as
// the conjunction of clauses.
type Clause interface {
	// clause is a marker interface to prevent external package from implementing
	// the interface.
	clause()
}

// newTriples declares that an attribute of the entity represented by this var
// should have the provided  value. Note that this introduces the var as an
// entity into the query if it is not already one.
func newTriple(entity Var, a Attr, value expr) Clause {
	return &tripleDecl{
		entity:    entity,
		attribute: a,
		value:     value,
	}
}

// AttrEq constrains the entity bound to v to have the provided value for
// the specified attr.
func (v Var) AttrEq(a Attr, value interface{}) Clause {
	return newTriple(v, a, valueExpr{value: value})
}

// AttrIn constrains the entity bound to v to have a value for
// the specified attr in the set of provided values.
func (v Var) AttrIn(a Attr, values ...interface{}) Clause {
	return newTriple(v, a, (anyExpr)(values))
}

// AttrEqVar constrains the entity bound to v to have a value for
// the specified attr equal to the variable value.
func (v Var) AttrEqVar(a Attr, value Var) Clause {
	return newTriple(v, a, value)
}

// Eq return a clause enforcing that the var is the value
// provided.
func (v Var) Eq(value interface{}) Clause {
	return &eqDecl{v, valueExpr{value: value}}
}

// In return a clause enforcing that the var is one of the value
// provided.
func (v Var) In(disjuncts ...interface{}) Clause {
	return &eqDecl{v, (anyExpr)(disjuncts)}
}

// Type returns a clause enforcing that the variable has one of the types
// passed by constraining its Type to the output of passing the
// args to Types. It is syntactic sugar around existing primitives.
func (v Var) Type(valueForTypeOf interface{}, moreValuesForTypeOf ...interface{}) Clause {
	typ := reflect.TypeOf(valueForTypeOf)
	if len(moreValuesForTypeOf) == 0 {
		return v.AttrEq(Type, typ)
	}

	types := make([]interface{}, 0, len(moreValuesForTypeOf)+1)
	types = append(types, typ)
	for _, v := range moreValuesForTypeOf {
		types = append(types, reflect.TypeOf(v))
	}
	return v.AttrIn(Type, types...)
}

// Entities is a shorthand for defining all the entities as having this
// variable as their value for this attribute.
//
// TODO(ajwerner): Better name.
func (v Var) Entities(attr Attr, entities ...Var) Clause {
	terms := make([]Clause, len(entities))
	for i, e := range entities {
		terms[i] = e.AttrEqVar(attr, v)
	}
	return And(terms...)
}

// And constructs a clause represents a set of clauses which should
// be taken in conjunction and exist so that go functions can be written to
// return a single clause without needing to get involved in appending to
// a slice of clauses.
func And(terms ...Clause) Clause {
	return (and)(terms)
}

// Filter is used to construct a clause which runs an arbitrary predicate
// over variables.
func Filter(name string, vars ...Var) func(predicateFunc interface{}) Clause {
	return func(predicateFunc interface{}) Clause {
		return &filterDecl{
			name:          name,
			vars:          vars,
			predicateFunc: predicateFunc,
		}
	}
}
