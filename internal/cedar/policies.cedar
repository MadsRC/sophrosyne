// Initial policy
permit (
    principal,
    action == Action::"Users::RotateToken",
    resource
) when {
    principal.is_admin == true
    ||
    principal.id == resource.id
};
// Do not allow rotating tokens unless user is admin or the token is theirs
forbid (
    principal,
    action == Action::"Users::RotateToken",
    resource
) when {
    principal.is_admin == false
    &&
    principal.id != resource.id
};
permit (
    principal,
    action,
    resource
) when {
    principal.is_admin == true
};
permit (
    principal,
    action == Action::"Users::*",
    resource
) when {
    principal.id == resource.id
};
