# Alembic configuration for database migrations
from logging.config import fileConfig
from sqlalchemy import engine_from_config, pool
from alembic import context
import os
import sys

# Add src to path
sys.path.insert(0, os.path.dirname(os.path.dirname(__file__)))

from src.core.infrastructure.database import Base
from src.config import settings

# Import all models so Alembic can detect them
from src.modules.auth.infrastructure.models.user_model import UserModel
from src.modules.stores.infrastructure.models.store_model import StoreModel, DeliveryPointModel
from src.modules.products.infrastructure.models.product_model import ProductModel
from src.modules.orders.infrastructure.models.order_model import OrderModel, OrderItemModel
from src.modules.payments.infrastructure.models.payment_model import PaymentModel, PlatformRevenueModel

# this is the Alembic Config object
config = context.config

# Interpret the config file for Python logging
if config.config_file_name is not None:
    fileConfig(config.config_file_name)

# Set target metadata
target_metadata = Base.metadata

# Override sqlalchemy.url from settings
config.set_main_option("sqlalchemy.url", settings.DATABASE_URL.replace("+asyncpg", ""))


def run_migrations_offline() -> None:
    """Run migrations in 'offline' mode."""
    url = config.get_main_option("sqlalchemy.url")
    context.configure(
        url=url,
        target_metadata=target_metadata,
        literal_binds=True,
        dialect_opts={"paramstyle": "named"},
    )

    with context.begin_transaction():
        context.run_migrations()


def run_migrations_online() -> None:
    """Run migrations in 'online' mode."""
    connectable = engine_from_config(
        config.get_section(config.config_ini_section, {}),
        prefix="sqlalchemy.",
        poolclass=pool.NullPool,
    )

    with connectable.connect() as connection:
        context.configure(
            connection=connection,
            target_metadata=target_metadata
        )

        with context.begin_transaction():
            context.run_migrations()


if context.is_offline_mode():
    run_migrations_offline()
else:
    run_migrations_online()
