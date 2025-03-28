box.cfg{
  listen = '0.0.0.0:3301'
}

box.schema.space.create('votes', {
  if_not_exists = true,
  format = {
      {name = 'id', type = 'string'},
      {name = 'creator_id', type = 'string'},
      {name = 'channel_id', type = 'string'},
      {name = 'question', type = 'string'},
      {name = 'options', type = 'array'},
      {name = 'votes', type = 'map'},
      {name = 'created_at', type = 'unsigned'},
      {name = 'is_active', type = 'boolean'}
  }
})

box.space.votes:create_index('primary', {
  type = 'hash',
  parts = {'id'},
  if_not_exists = true
})