class DomainException(Exception):
    """Base domain exception."""
    pass


class EntityNotFoundException(DomainException):
    """Raised when an entity is not found."""
    
    def __init__(self, entity_name: str, entity_id: int):
        self.entity_name = entity_name
        self.entity_id = entity_id
        super().__init__(f"{entity_name} with id {entity_id} not found")


class ValidationException(DomainException):
    """Raised when domain validation fails."""
    pass


class UnauthorizedException(DomainException):
    """Raised when user lacks permission."""
    pass


class BusinessRuleViolationException(DomainException):
    """Raised when a business rule is violated."""
    pass
