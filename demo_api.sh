#!/bin/bash

API_URL="http://localhost:8080/api/v1"
TOKEN=""
DEMO_MODE="${1:-false}"

echo "=========================================="
echo "      Todo List API 演示"
echo "=========================================="

if [ "$DEMO_MODE" = "true" ]; then
  echo "Demo模式：每次API调用后sleep 3秒"
fi

# 1. 健康检查
echo ""
echo "1. 健康检查"
echo "----------------------------------------"
echo "URI: /health"
echo "curl -s http://localhost:8080/health | jq ."
curl -s http://localhost:8080/health | jq .
[ "$DEMO_MODE" = "true" ] && sleep 3

# 2. 用户登录
echo ""
echo "2. 用户登录"
echo "----------------------------------------"
echo "URI: /api/v1/auth/login"
echo "curl -s -X POST $API_URL/auth/login -H \"Content-Type: application/json\" -d '{\"username\":\"admin\",\"password\":\"admin\"}'"
LOGIN_RESPONSE=$(curl -s -X POST $API_URL/auth/login -H "Content-Type: application/json" -d '{"username":"admin","password":"admin"}')
echo "$LOGIN_RESPONSE" | jq .
TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.access_token')

# 检查登录是否成功
if [ "$TOKEN" == "null" ] || [ -z "$TOKEN" ]; then
  echo ""
  echo "登录失败！无法获取 access_token，终止后续流程"
  echo ""
  echo "=========================================="
  echo "      演示失败！"
  echo "=========================================="
  exit 1
fi

echo ""
echo "Access Token 已获取（前20字符）: ${TOKEN:0:20}..."
[ "$DEMO_MODE" = "true" ] && sleep 3

# 3. 创建 Todo
echo ""
echo "3. 创建 Todo"
echo "----------------------------------------"
echo "URI: /api/v1/todos"
echo "curl -s -X POST \$API_URL/todos -H \"Content-Type: application/json\" -H \"Authorization: Bearer \$TOKEN\" -d '{\"title\":\"面试演示：学习 Go Clean Architecture\",\"description\":\"深入学习 Clean Architecture 的最佳实践，包括依赖倒置、单向依赖、框架无关等核心概念\",\"priority\":\"high\",\"due_date\":\"2026-02-15T10:00:00Z\",\"tags\":[\"学习\",\"Go\",\"架构\"]}'"
curl -s -X POST $API_URL/todos -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"title":"面试演示：学习 Go Clean Architecture","description":"深入学习 Clean Architecture 的最佳实践，包括依赖倒置、单向依赖、框架无关等核心概念","priority":"high","due_date":"2026-02-15T10:00:00Z","tags":["学习","Go","架构"]}' | jq .
[ "$DEMO_MODE" = "true" ] && sleep 3

# 4. 查询 Todo 列表
echo ""
echo "4. 查询 Todo 列表"
echo "----------------------------------------"
echo "URI: /api/v1/todos?page=1&limit=20&sort_by=due_date&sort_order=asc"
echo "curl -s -X GET \"$API_URL/todos?page=1&limit=20&sort_by=due_date&sort_order=asc\" -H \"Authorization: Bearer \$TOKEN\""
curl -s -X GET "$API_URL/todos?page=1&limit=20&sort_by=due_date&sort_order=asc" -H "Authorization: Bearer $TOKEN" | jq .
[ "$DEMO_MODE" = "true" ] && sleep 3

# 5. 查询单个 Todo
echo ""
echo "5. 查询单个 Todo (ID=1)"
echo "----------------------------------------"
echo "URI: /api/v1/todos/1"
echo "curl -s -X GET \"$API_URL/todos/1\" -H \"Authorization: Bearer \$TOKEN\""
curl -s -X GET "$API_URL/todos/1" -H "Authorization: Bearer $TOKEN" | jq .
[ "$DEMO_MODE" = "true" ] && sleep 3

# 6. 更新 Todo 状态
echo ""
echo "6. 更新 Todo 状态"
echo "----------------------------------------"
echo "URI: /api/v1/todos/1/status"
echo "curl -s -X PATCH \"$API_URL/todos/1/status\" -H \"Content-Type: application/json\" -H \"Authorization: Bearer \$TOKEN\" -d '{\"status\":\"in_progress\"}'"
curl -s -X PATCH "$API_URL/todos/1/status" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"status":"in_progress"}' | jq .
[ "$DEMO_MODE" = "true" ] && sleep 3

# 7. 按状态过滤
echo ""
echo "7. 按状态过滤（in_progress）"
echo "----------------------------------------"
echo "URI: /api/v1/todos?status=in_progress&page=1&limit=20"
echo "curl -s -X GET \"$API_URL/todos?status=in_progress&page=1&limit=20\" -H \"Authorization: Bearer \$TOKEN\""
curl -s -X GET "$API_URL/todos?status=in_progress&page=1&limit=20" -H "Authorization: Bearer $TOKEN" | jq .
[ "$DEMO_MODE" = "true" ] && sleep 3

# 8. 按优先级过滤
echo ""
echo "8. 按优先级过滤（high）"
echo "----------------------------------------"
echo "URI: /api/v1/todos?priority=high&page=1&limit=20"
echo "curl -s -X GET \"$API_URL/todos?priority=high&page=1&limit=20\" -H \"Authorization: Bearer \$TOKEN\""
curl -s -X GET "$API_URL/todos?priority=high&page=1&limit=20" -H "Authorization: Bearer $TOKEN" | jq .
[ "$DEMO_MODE" = "true" ] && sleep 3

# 9. 搜索功能
echo ""
echo "9. 搜索功能（关键词：Go）"
echo "----------------------------------------"
echo "URI: /api/v1/todos?search=Go&page=1&limit=20"
echo "curl -s -X GET \"$API_URL/todos?search=Go&page=1&limit=20\" -H \"Authorization: Bearer \$TOKEN\""
curl -s -X GET "$API_URL/todos?search=Go&page=1&limit=20" -H "Authorization: Bearer $TOKEN" | jq .
[ "$DEMO_MODE" = "true" ] && sleep 3

# 10. 创建更多 Todo（展示分页）
echo ""
echo "10. 创建更多 Todo（展示分页）"
echo "----------------------------------------"
echo "URI: /api/v1/todos"
for i in 2 3 4; do
  echo "curl -s -X POST \$API_URL/todos -H \"Content-Type: application/json\" -H \"Authorization: Bearer \$TOKEN\" -d '{\"title\":\"Todo $i - 演示分页功能\",\"description\":\"这是第 $i 个 Todo\",\"priority\":\"medium\",\"tags\":[\"演示\"]}'"
  curl -s -X POST $API_URL/todos -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d "{\"title\":\"Todo $i - 演示分页功能\",\"description\":\"这是第 $i 个 Todo\",\"priority\":\"medium\",\"tags\":[\"演示\"]}" > /dev/null
done
echo "已创建 3 个额外 Todo"
[ "$DEMO_MODE" = "true" ] && sleep 3

# 11. 分页查询
echo ""
echo "11. 分页查询（第1页，每页2个）"
echo "----------------------------------------"
echo "URI: /api/v1/todos?page=1&limit=2"
echo "curl -s -X GET \"$API_URL/todos?page=1&limit=2\" -H \"Authorization: Bearer \$TOKEN\" | jq '.data[] | {id, title, status}'"
curl -s -X GET "$API_URL/todos?page=1&limit=2" -H "Authorization: Bearer $TOKEN" | jq '.data[] | {id, title, status}'
[ "$DEMO_MODE" = "true" ] && sleep 3

# 12. 第2页
echo ""
echo "12. 分页查询（第2页，每页2个）"
echo "----------------------------------------"
echo "URI: /api/v1/todos?page=2&limit=2"
echo "curl -s -X GET \"$API_URL/todos?page=2&limit=2\" -H \"Authorization: Bearer \$TOKEN\" | jq '.data[] | {id, title, status}'"
curl -s -X GET "$API_URL/todos?page=2&limit=2" -H "Authorization: Bearer $TOKEN" | jq '.data[] | {id, title, status}'
[ "$DEMO_MODE" = "true" ] && sleep 3

# 13. 查看 Tags
echo ""
echo "13. 查看所有 Tags"
echo "----------------------------------------"
echo "URI: /api/v1/tags?page=1&limit=20"
echo "curl -s -X GET \"$API_URL/tags?page=1&limit=20\" -H \"Authorization: Bearer \$TOKEN\""
curl -s -X GET "$API_URL/tags?page=1&limit=20" -H "Authorization: Bearer $TOKEN" | jq .
[ "$DEMO_MODE" = "true" ] && sleep 3

# 14. 查看用户的 Tags
echo ""
echo "14. 查看用户的 Tags"
echo "----------------------------------------"
echo "URI: /api/v1/users/my-tags"
echo "curl -s -X GET \"$API_URL/users/my-tags\" -H \"Authorization: Bearer \$TOKEN\""
curl -s -X GET "$API_URL/users/my-tags" -H "Authorization: Bearer $TOKEN" | jq .
[ "$DEMO_MODE" = "true" ] && sleep 3

# 15. 更新 Todo
echo ""
echo "15. 更新 Todo（修改标题和描述）"
echo "----------------------------------------"
echo "URI: /api/v1/todos/1"
echo "curl -s -X PUT \"$API_URL/todos/1\" -H \"Content-Type: application/json\" -H \"Authorization: Bearer \$TOKEN\" -d '{\"title\":\"面试演示：学习 Go Clean Architecture（已更新）\",\"description\":\"深入学习 Clean Architecture，包括依赖倒置、单向依赖、框架无关等核心概念。已理解并实践！\"}'"
curl -s -X PUT "$API_URL/todos/1" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"title":"面试演示：学习 Go Clean Architecture（已更新）","description":"深入学习 Clean Architecture，包括依赖倒置、单向依赖、框架无关等核心概念。已理解并实践！"}' | jq .
[ "$DEMO_MODE" = "true" ] && sleep 3

# 16. 完成 Todo
echo ""
echo "16. 完成 Todo"
echo "----------------------------------------"
echo "URI: /api/v1/todos/1/status"
echo "curl -s -X PATCH \"$API_URL/todos/1/status\" -H \"Content-Type: application/json\" -H \"Authorization: Bearer \$TOKEN\" -d '{\"status\":\"completed\"}'"
curl -s -X PATCH "$API_URL/todos/1/status" -H "Content-Type: application/json" -H "Authorization: Bearer $TOKEN" -d '{"status":"completed"}' | jq .
[ "$DEMO_MODE" = "true" ] && sleep 3

# 17. 查看已完成的 Todos
echo ""
echo "17. 查看已完成的 Todos"
echo "----------------------------------------"
echo "URI: /api/v1/todos?status=completed&page=1&limit=20"
echo "curl -s -X GET \"$API_URL/todos?status=completed&page=1&limit=20\" -H \"Authorization: Bearer \$TOKEN\" | jq '.data[] | {id, title, status}'"
curl -s -X GET "$API_URL/todos?status=completed&page=1&limit=20" -H "Authorization: Bearer $TOKEN" | jq '.data[] | {id, title, status}'
[ "$DEMO_MODE" = "true" ] && sleep 3

# 18. 用户登出
echo ""
echo "18. 用户登出"
echo "----------------------------------------"
echo "URI: /api/v1/auth/logout"
echo "curl -s -X POST $API_URL/auth/logout -H \"Authorization: Bearer \$TOKEN\""
curl -s -X POST $API_URL/auth/logout -H "Authorization: Bearer $TOKEN" | jq .

echo ""
echo "=========================================="
echo "      演示完成！"
echo "=========================================="
