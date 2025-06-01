<template>
  <div class="chat-area">
    <!-- 聊天内容 -->
    <div class="chat-content">
      <!-- 聊天消息列表 -->
      <ul class="message-list">
        <!-- 消息项 -->
        <li v-for="(message, index) in messages" :key="index" class="message-item">
          {{ message }}
        </li>
      </ul>

      <!-- 输入框 -->
      <div class="input-box">
        <input type="text" v-model="newMessage" @keyup.enter="sendMessage" placeholder="请输入消息..." />
        <button @click="sendMessage">发送</button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'

const messages = ref(['欢迎使用聊天系统！'])
const newMessage = ref('')

const sendMessage = () => {
  if (newMessage.value.trim() !== '') {
    messages.value.push(newMessage.value)
    newMessage.value = ''
  }
}
</script>

<style scoped>
.chat-area {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.chat-content {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.message-list {
  list-style: none;
  padding: 0;
  margin: 0;
  flex: 1;
  overflow-y: auto;
}

.message-item {
  padding: 8px;
  border-bottom: 1px solid #eee;
}

.input-box {
  display: flex;
  align-items: center;
  padding: 8px;
  border-top: 1px solid #eee;
}

.input-box input {
  flex: 1;
  padding: 8px;
  border: 1px solid #ddd;
  border-radius: 4px;
  margin-right: 8px;
}

.input-box button {
  padding: 8px 16px;
  border: none;
  background-color: #007bff;
  color: white;
  border-radius: 4px;
  cursor: pointer;
}
</style>