Create TYPE delivery_state As enum ('searching','coming', 'picked', 'dropped');
CREATE TABLE IF NOT EXISTS orders (
                                      id bigserial PRIMARY KEY,
                                      created_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                      updated_at timestamp(0) WITH TIME ZONE NOT NULL DEFAULT NOW(),
                                      finishCode numeric(4,0) DEFAULT (FLOOR(random() * 9999) + 1)::numeric,
                                      CHECK (finishCode >= 1 AND finishCode <= 9999),
                                      isCompleted bool NOT NULL DEFAULT false,
                                      originLatitude numeric NOT NULL,
                                      originLongitude numeric NOT NULL,
                                      destinationLatitude numeric NOT NULL,
                                      destinationLongitude numeric NOT NULL,
                                      deliveryId bigint REFERENCES users,
                                      customerId bigint NOT NULL REFERENCES users,
                                      delivery delivery_state NOT NULL default 'searching',
                                      origin_adress varchar(70) not null ,
                                      destination_adress varchar(70) not null ,
                                      version integer NOT NULL DEFAULT 1
);
