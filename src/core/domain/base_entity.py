from datetime import datetime
from typing import Optional


class BaseEntity:
    """Base domain entity with common attributes."""
    
    id: Optional[int] = None
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
