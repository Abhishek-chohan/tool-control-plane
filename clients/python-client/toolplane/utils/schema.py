import inspect
from typing import Any, Dict, List, Optional, Union

from typing_extensions import get_args, get_origin, get_type_hints

__all__ = ["generate_schema_from_function", "type_to_json_schema"]


def generate_schema_from_function(func):
    """
    Generate a JSON schema for a function based on type hints and docstring.

    Args:
        func (Callable): Function to generate schema for

    Returns:
        Dict: JSON schema for the function
    """
    sig = inspect.signature(func)
    type_hints = get_type_hints(func)

    # Parse the docstring to get parameter descriptions
    docstring = inspect.getdoc(func) or ""
    param_descriptions = {}
    description = ""

    lines = docstring.split("\n")
    in_params = False
    current_param = None

    for line in lines:
        line = line.strip()
        if not line:
            continue

        if line.lower().startswith("parameters:") or line.lower().startswith("args:"):
            in_params = True
            continue

        if in_params:
            # Check if this line defines a new parameter
            if ":" in line and line[0].isalnum():
                parts = line.split(":", 1)
                current_param = parts[0].strip()
                param_descriptions[current_param] = parts[1].strip()
            elif current_param and line[0].isspace():
                # Continuation of previous parameter description
                param_descriptions[current_param] += " " + line.strip()
        elif not description:
            # First line(s) before parameters is the description
            description += line + " "

    # Build the schema
    properties = {}
    required = []

    for param_name, param in sig.parameters.items():
        if param_name == "self":
            continue

        param_type = type_hints.get(param_name, Any)
        param_schema = type_to_json_schema(param_type)

        # Add description if available
        if param_name in param_descriptions:
            param_schema["description"] = param_descriptions[param_name]

        properties[param_name] = param_schema

        # If the parameter has no default value, it's required
        if param.default is inspect.Parameter.empty:
            required.append(param_name)

    schema = {"type": "object", "properties": properties}

    if required:
        schema["required"] = required

    return {
        "name": func.__name__,
        "description": description.strip(),
        "schema": schema,
    }


def type_to_json_schema(type_hint):
    """
    Convert a Python type hint to a JSON schema type.

    Args:
        type_hint: Python type hint

    Returns:
        Dict: JSON schema type definition
    """
    origin = get_origin(type_hint)
    args = get_args(type_hint)

    # Handle Union types (Optional is Union[type, None])
    if origin is Union:
        if len(args) == 2 and type(None) in args:
            # Handle Optional[type]
            for arg in args:
                if arg is not type(None):
                    schema = type_to_json_schema(arg)
                    schema["nullable"] = True
                    return schema
        else:
            # General Union - not fully supported in JSON Schema
            return {"type": "object"}

    # Handle List types
    if origin is list or origin is List:
        if args:
            return {"type": "array", "items": type_to_json_schema(args[0])}
        else:
            return {"type": "array"}

    # Handle Dict types
    if origin is dict or origin is Dict:
        return {"type": "object"}

    # Handle primitive types
    if type_hint is str or type_hint is Optional[str]:
        return {"type": "string"}
    elif type_hint is int or type_hint is Optional[int]:
        return {"type": "integer"}
    elif type_hint is float or type_hint is Optional[float]:
        return {"type": "number"}
    elif type_hint is bool or type_hint is Optional[bool]:
        return {"type": "boolean"}
    elif (
        type_hint is list
        or type_hint is List
        or type_hint is Optional[list]
        or type_hint is Optional[List]
    ):
        return {"type": "array"}
    elif (
        type_hint is dict
        or type_hint is Dict
        or type_hint is Optional[dict]
        or type_hint is Optional[Dict]
    ):
        return {"type": "object"}  # Default to object for complex types
    return {"type": "object"}
